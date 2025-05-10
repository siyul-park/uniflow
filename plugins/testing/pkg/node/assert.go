package node

import (
	"context"
	"fmt"

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
	agent   *runtime.Agent
	spec    *AssertNodeSpec
	fn      func(context.Context, interface{}) (bool, error)
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

		evaluator := func(ctx context.Context, payload interface{}) (bool, error) {
			result, err := program.Run(ctx, payload)
			if err != nil {
				return false, err
			}

			boolResult, ok := result.(bool)
			if !ok {
				return false, fmt.Errorf("expression must evaluate to a boolean, got %T for '%s'",
					result, spec.Expect)
			}

			return boolResult, nil
		}

		return NewAssertNode(spec, agent, evaluator), nil
	})
}

// NewAssertNode creates a new Assert node with the given spec and agent
func NewAssertNode(spec *AssertNodeSpec, agent *runtime.Agent, fn func(context.Context, interface{}) (bool, error)) *AssertNode {
	n := &AssertNode{
		inPort:  port.NewIn(),
		outPort: port.NewOut(),
		agent:   agent,
		spec:    spec,
		fn:      fn,
	}

	n.inPort.AddListener(port.ListenFunc(n.forward))

	return n
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

	var target interface{}

	for inPck := range inReader.Read() {
		value, ok := inPck.Payload().(types.Slice)
		if !ok || value.Len() != 2 {
			inReader.Receive(packet.New(types.NewError(fmt.Errorf("invalid packet format"))))
			continue
		}

		inPayload, frameIdx := value.Get(0), value.Get(1)

		if n.spec.Target != nil {
			frame, err := n.find(n.spec.Target)
			if err != nil {
				inReader.Receive(packet.New(types.NewError(err)))
				continue
			}
			target = frame
		} else {
			target = inPayload
		}

		result, err := n.fn(proc, target)
		if err != nil {
			inReader.Receive(packet.New(types.NewError(err)))
			continue
		}

		if !result {
			errMsg := fmt.Errorf("assertion failed: expression '%s' evaluated to false with value %v",
				n.spec.Expect, target)
			inReader.Receive(packet.New(types.NewError(errMsg)))
			continue
		}

		outPayload, err := types.Marshal([]interface{}{inPayload, frameIdx})
		if err != nil {
			inReader.Receive(packet.New(types.NewError(err)))
			continue
		}

		outPck := packet.New(outPayload)
		outWriter.Write(outPck)
		inReader.Receive(packet.None)
	}
}

// find locates the target frame based on the target specification
func (n *AssertNode) find(target *AssertNodeTarget) (interface{}, error) {
	if target == nil {
		return nil, fmt.Errorf("no target specified")
	}

	tname := target.Name
	tport := target.Port

	for _, proc := range n.agent.Processes() {
		for _, frm := range n.agent.Frames(proc.ID()) {
			if frm.Symbol == nil {
				continue
			}

			if frm.Symbol.Name() != tname {
				continue
			}

			if inPort := frm.Symbol.In(tport); inPort != nil && frm.InPck != nil {
				return frm.InPck.Payload(), nil
			}

			if outPort := frm.Symbol.Out(tport); outPort != nil && frm.OutPck != nil {
				return frm.OutPck.Payload(), nil
			}
		}
	}

	return nil, fmt.Errorf("target frame not found: node '%s' with port '%s' could not be located",
		tname, tport)
}
