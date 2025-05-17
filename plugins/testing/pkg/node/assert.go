package node

import (
	"context"
	"fmt"
	"time"

	"github.com/siyul-park/uniflow/pkg/language"
	"github.com/siyul-park/uniflow/pkg/node"
	"github.com/siyul-park/uniflow/pkg/packet"
	"github.com/siyul-park/uniflow/pkg/port"
	"github.com/siyul-park/uniflow/pkg/process"
	"github.com/siyul-park/uniflow/pkg/runtime"
	"github.com/siyul-park/uniflow/pkg/scheme"
	"github.com/siyul-park/uniflow/pkg/spec"
	"github.com/siyul-park/uniflow/pkg/types"
)

// AssertNodeSpec defines the specification for Assert node
type AssertNodeSpec struct {
	spec.Meta `json:",inline"`
	Expect    string            `json:"expect"`
	Target    *AssertNodeTarget `json:"target,omitempty"`
	Timeout   time.Duration     `json:"timeout,omitempty"`
}

// AssertNodeTarget defines the target to validate
type AssertNodeTarget struct {
	Name string `json:"name"`
	Port string `json:"port"`
}

// AssertNode implements the Assert node functionality
type AssertNode struct {
	inPort  *port.InPort
	outPort *port.OutPort
	expect  func(context.Context, interface{}) (bool, error)
	target  func(interface{}, interface{}) (interface{}, interface{}, error)
	agent   *runtime.Agent
}

const KindAssert = "assert"

var _ node.Node = (*AssertNode)(nil)

// NewAssertNodeCodec creates a codec for AssertNode
func NewAssertNodeCodec(agent *runtime.Agent, compiler language.Compiler) scheme.Codec {
	return scheme.CodecWithType(func(spec *AssertNodeSpec) (node.Node, error) {
		program, err := compiler.Compile(spec.Expect)
		if err != nil {
			return nil, err
		}

		evaluator := language.Predicate[any](language.Timeout(program, spec.Timeout))

		n := NewAssertNode()
		n.SetExpect(evaluator)
		n.SetAgent(agent)

		if spec.Target != nil {
			n.SetTarget(spec.Target.Name, spec.Target.Port)
		}

		return n, nil
	})
}

// NewAssertNode creates a new Assert node
func NewAssertNode() *AssertNode {
	n := &AssertNode{
		inPort:  port.NewIn(),
		outPort: port.NewOut(),
		target: func(payload interface{}, index interface{}) (interface{}, interface{}, error) {
			return payload, index, nil
		},
	}

	n.inPort.AddListener(port.ListenFunc(n.forward))
	return n
}

// SetExpect sets the expectation function
func (n *AssertNode) SetExpect(expect func(context.Context, interface{}) (bool, error)) {
	n.expect = expect
}

// SetTarget sets the target function
func (n *AssertNode) SetTarget(name string, port string) {
	n.target = func(payload interface{}, index interface{}) (interface{}, interface{}, error) {
		if n.agent == nil {
			return nil, nil, fmt.Errorf("agent not set")
		}

		target, err := find(n.agent, name, port)
		if err != nil {
			return nil, nil, err
		}
		return target, index, nil
	}
}

// SetAgent sets the runtime agent for finding targets
func (n *AssertNode) SetAgent(agent *runtime.Agent) {
	n.agent = agent
}

// In returns the input port with the specified name
func (n *AssertNode) In(name string) *port.InPort {
	if name == "in" {
		return n.inPort
	}
	return nil
}

// Out returns the output port with the specified name
func (n *AssertNode) Out(name string) *port.OutPort {
	if name == "out" {
		return n.outPort
	}
	return nil
}

// Close closes the ports
func (n *AssertNode) Close() error {
	n.inPort.Close()
	n.outPort.Close()
	return nil
}

func (n *AssertNode) forward(proc *process.Process) {
	inReader := n.inPort.Open(proc)
	outWriter := n.outPort.Open(proc)

	for inPck := range inReader.Read() {
		value, ok := inPck.Payload().(types.Slice)
		if !ok || value.Len() != 2 {
			inReader.Receive(packet.New(types.NewError(fmt.Errorf("invalid packet format"))))
			continue
		}

		inPayload, frameIdx := value.Get(0), value.Get(1)

		target, idx, err := n.target(inPayload, frameIdx)
		if err != nil {
			inReader.Receive(packet.New(types.NewError(err)))
			continue
		}

		ok, err = n.expect(proc, target)
		if err != nil {
			inReader.Receive(packet.New(types.NewError(err)))
			continue
		}

		if !ok {
			inReader.Receive(packet.New(types.NewError(fmt.Errorf("assertion failed: evaluated to false with value %v", target))))
			continue
		}

		outPayload, err := types.Marshal([]interface{}{inPayload, idx})
		if err != nil {
			inReader.Receive(packet.New(types.NewError(err)))
			continue
		}

		outWriter.Write(packet.New(outPayload))
		inReader.Receive(packet.None)
	}
}

func find(agent *runtime.Agent, name string, port string) (interface{}, error) {
	for _, proc := range agent.Processes() {
		for _, frm := range agent.Frames(proc.ID()) {
			if frm.Symbol == nil || frm.Symbol.Name() != name {
				continue
			}

			if inPort := frm.Symbol.In(port); inPort != nil && frm.InPck != nil {
				return frm.InPck.Payload(), nil
			}

			if outPort := frm.Symbol.Out(port); outPort != nil && frm.OutPck != nil {
				return frm.OutPck.Payload(), nil
			}
		}
	}

	return nil, fmt.Errorf("target frame not found: node '%s' with port '%s' could not be located",
		name, port)
}
