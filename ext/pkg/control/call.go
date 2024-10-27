package control

import (
	"github.com/siyul-park/uniflow/pkg/node"
	"github.com/siyul-park/uniflow/pkg/packet"
	"github.com/siyul-park/uniflow/pkg/port"
	"github.com/siyul-park/uniflow/pkg/process"
	"github.com/siyul-park/uniflow/pkg/scheme"
	"github.com/siyul-park/uniflow/pkg/spec"
	"github.com/siyul-park/uniflow/pkg/types"
)

// CallNodeSpec holds the specification for creating a CallNode.
type CallNodeSpec struct {
	spec.Meta `map:",inline"`
}

// CallNode processes an input packet and sends the result to multiple output ports.
type CallNode struct {
	tracer   *packet.Tracer
	inPort   *port.InPort
	outPorts []*port.OutPort
	errPort  *port.OutPort
}

const KindCall = "call"

var _ node.Node = (*CallNode)(nil)

// NewCallNodeCodec creates a new codec for CallNodeSpec.
func NewCallNodeCodec() scheme.Codec {
	return scheme.CodecWithType(func(spec *CallNodeSpec) (node.Node, error) {
		return NewCallNode(), nil
	})
}

// NewCallNode creates a new CallNode.
func NewCallNode() *CallNode {
	n := &CallNode{
		tracer:   packet.NewTracer(),
		inPort:   port.NewIn(),
		outPorts: []*port.OutPort{port.NewOut(), port.NewOut()},
		errPort:  port.NewOut(),
	}

	n.inPort.AddListener(port.ListenFunc(n.forward))
	n.outPorts[0].AddListener(n.backward(0))
	n.outPorts[1].AddListener(n.backward(1))
	n.errPort.AddListener(port.ListenFunc(n.catch))

	return n
}

// In returns the input port with the specified name.
func (n *CallNode) In(name string) *port.InPort {
	switch name {
	case node.PortIn:
		return n.inPort
	default:
		return nil
	}
}

// Out returns the output port with the specified name.
func (n *CallNode) Out(name string) *port.OutPort {
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
func (n *CallNode) Close() error {
	n.inPort.Close()
	for _, outPort := range n.outPorts {
		outPort.Close()
	}
	n.errPort.Close()
	n.tracer.Close()
	return nil
}

func (n *CallNode) forward(proc *process.Process) {
	inReader := n.inPort.Open(proc)
	outWriter0 := n.outPorts[0].Open(proc)
	outWriter1 := n.outPorts[1].Open(proc)
	errWriter := n.errPort.Open(proc)

	for inPck := range inReader.Read() {
		n.tracer.Read(inReader, inPck)

		n.tracer.AddHook(inPck, packet.HookFunc(func(backPck *packet.Packet) {
			n.tracer.Transform(inPck, backPck)
			if _, ok := backPck.Payload().(types.Error); ok {
				n.tracer.Write(errWriter, backPck)
			} else {
				n.tracer.Write(outWriter1, backPck)
			}
		}))

		n.tracer.Write(outWriter0, inPck)
	}
}

func (n *CallNode) backward(index int) port.Listener {
	return port.ListenFunc(func(proc *process.Process) {
		outWriter := n.outPorts[index].Open(proc)

		for backPck := range outWriter.Receive() {
			n.tracer.Receive(outWriter, backPck)
		}
	})
}

func (n *CallNode) catch(proc *process.Process) {
	errWriter := n.errPort.Open(proc)

	for backPck := range errWriter.Receive() {
		n.tracer.Receive(errWriter, backPck)
	}
}
