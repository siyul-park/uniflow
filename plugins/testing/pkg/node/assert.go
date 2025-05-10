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
	inPort    *port.InPort
	outPort   *port.OutPort
	agent     *runtime.Agent
	spec      *AssertNodeSpec
	condition language.Program
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

		return NewAssertNode(spec, agent, program), nil
	})
}

// NewAssertNode creates a new Assert node with the given spec and agent
func NewAssertNode(spec *AssertNodeSpec, agent *runtime.Agent, condition language.Program) *AssertNode {
	n := &AssertNode{
		inPort:    port.NewIn(),
		outPort:   port.NewOut(),
		agent:     agent,
		spec:      spec,
		condition: condition,
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
		value, ok := types.InterfaceOf(inPck.Payload()).([]interface{})
		if !ok || len(value) != 2 {
			inReader.Receive(packet.New(types.NewError(fmt.Errorf("invalid packet format"))))
			continue
		}

		inPayload, frameIdx := value[0], value[1]

		if n.spec.Target != nil {
			frm, err := n.findTargetFrame()
			if err != nil {
				inReader.Receive(packet.New(types.NewError(err)))
				continue
			}
			target = frm
		} else {
			target = inPayload
		}

		result, err := n.evaluate(proc, target)
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

func (n *AssertNode) findTargetFrame() (interface{}, error) {
	targetName := n.spec.Target.Name
	targetPort := n.spec.Target.Port

	for _, proc := range n.agent.Processes() {
		for _, frm := range n.agent.Frames(proc.ID()) {
			if frm.Symbol == nil {
				continue
			}

			if frm.Symbol.Name() != targetName {
				continue
			}

			if node.NameOfPort(targetPort) == node.PortIn {
				for name := range frm.Symbol.Ins() {
					if name == targetPort {
						return types.InterfaceOf(frm.InPck.Payload()), nil
					}
				}
			}

			if node.NameOfPort(targetPort) == node.PortOut {
				for name := range frm.Symbol.Outs() {
					if name == targetPort {
						return types.InterfaceOf(frm.OutPck.Payload()), nil
					}
				}
			}
		}
	}

	return nil, fmt.Errorf("target frame not found: node '%s' with port '%s' could not be located",
		targetName, targetPort)
}

func (n *AssertNode) evaluate(ctx context.Context, payload interface{}) (bool, error) {
	result, err := n.condition.Run(ctx, payload)
	if err != nil {
		return false, fmt.Errorf("expression evaluation error: '%s' with payload %v: %w",
			n.spec.Expect, payload, err)
	}

	boolResult, ok := result.(bool)
	if !ok {
		return false, fmt.Errorf("expression must evaluate to a boolean, got %T for '%s'",
			result, n.spec.Expect)
	}

	return boolResult, nil
}
