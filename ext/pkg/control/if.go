package control

import (
	"reflect"

	"github.com/siyul-park/uniflow/ext/pkg/language"
	"github.com/siyul-park/uniflow/pkg/node"
	"github.com/siyul-park/uniflow/pkg/packet"
	"github.com/siyul-park/uniflow/pkg/port"
	"github.com/siyul-park/uniflow/pkg/process"
	"github.com/siyul-park/uniflow/pkg/scheme"
	"github.com/siyul-park/uniflow/pkg/spec"
	"github.com/siyul-park/uniflow/pkg/types"
)

// IfNodeSpec holds specifications for creating an IfNode.
type IfNodeSpec struct {
	spec.Meta `map:",inline"`
	When      string `map:"when"`
}

// IfNode represents a node that evaluates a condition and routes packets based on the result.
type IfNode struct {
	condition func(any) (bool, error)
	tracer    *packet.Tracer
	inPort    *port.InPort
	outPorts  []*port.OutPort
	errPort   *port.OutPort
}

const KindIf = "if"

// NewIfNodeCodec creates a new codec for IfNodeSpec.
func NewIfNodeCodec(compiler language.Compiler) scheme.Codec {
	return scheme.CodecWithType(func(spec *IfNodeSpec) (node.Node, error) {
		program, err := compiler.Compile(spec.When)
		if err != nil {
			return nil, err
		}
		return NewIfNode(func(env any) (bool, error) {
			res, err := program.Run(env)
			if err != nil {
				return false, err
			}
			return !reflect.ValueOf(res).IsZero(), nil
		}), nil
	})
}

// NewIfNode creates a new IfNode instance.
func NewIfNode(condition func(any) (bool, error)) *IfNode {
	n := &IfNode{
		condition: condition,
		tracer:    packet.NewTracer(),
		inPort:    port.NewIn(),
		outPorts:  []*port.OutPort{port.NewOut(), port.NewOut()},
		errPort:   port.NewOut(),
	}

	n.inPort.AddListener(port.ListenFunc(n.forward))

	for i, outPort := range n.outPorts {
		i := i
		outPort.AddListener(port.ListenFunc(func(proc *process.Process) {
			n.backward(proc, i)
		}))
	}

	n.errPort.AddListener(port.ListenFunc(n.catch))

	return n
}

// In returns the input port with the specified name.
func (n *IfNode) In(name string) *port.InPort {
	switch name {
	case node.PortIn:
		return n.inPort
	default:
	}
	return nil
}

// Out returns the output port with the specified name.
func (n *IfNode) Out(name string) *port.OutPort {
	switch name {
	case node.PortOut:
		return n.outPorts[0]
	case node.PortErr:
		return n.errPort
	default:
		if node.NameOfPort(name) == node.PortOut {
			index, ok := node.IndexOfPort(name)
			if ok && index < len(n.outPorts) {
				return n.outPorts[index]
			}
		}
	}
	return nil
}

// Close closes all ports associated with the node.
func (n *IfNode) Close() error {
	n.inPort.Close()
	for _, outPort := range n.outPorts {
		outPort.Close()
	}
	n.errPort.Close()
	n.tracer.Close()

	return nil
}

func (n *IfNode) forward(proc *process.Process) {
	inReader := n.inPort.Open(proc)
	outWriter0 := n.outPorts[0].Open(proc)
	outWriter1 := n.outPorts[1].Open(proc)
	errWriter := n.errPort.Open(proc)

	for {
		inPck, ok := <-inReader.Read()
		if !ok {
			return
		}
		n.tracer.Read(inReader, inPck)

		inPayload := inPck.Payload()
		input := types.InterfaceOf(inPayload)

		if ok, err := n.condition(input); err != nil {
			errPck := packet.New(types.NewError(err))
			n.tracer.Transform(inPck, errPck)
			n.tracer.Write(errWriter, errPck)
		} else if ok {
			n.tracer.Write(outWriter0, inPck)
		} else {
			n.tracer.Write(outWriter1, inPck)
		}
	}
}

func (n *IfNode) backward(proc *process.Process, index int) {
	outWriter := n.outPorts[index].Open(proc)

	for {
		backPck, ok := <-outWriter.Receive()
		if !ok {
			return
		}

		n.tracer.Receive(outWriter, backPck)
	}
}

func (n *IfNode) catch(proc *process.Process) {
	errWriter := n.errPort.Open(proc)

	for {
		backPck, ok := <-errWriter.Receive()
		if !ok {
			return
		}

		n.tracer.Receive(errWriter, backPck)
	}
}
