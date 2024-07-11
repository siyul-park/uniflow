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

// CallNode redirects packets from the input port to the intermediate port for processing by connected nodes, then outputs the results to the output port.
type CallNode struct {
	tracer   *packet.Tracer
	inPort   *port.InPort
	outPorts []*port.OutPort
	errPort  *port.OutPort
}

// CallNodeSpec holds the specifications for creating a CallNode.
type CallNodeSpec struct {
	spec.Meta `map:",inline"`
}

const KindCall = "call"

var _ node.Node = (*CallNode)(nil)

// NewCallNode creates a new CallNode.
func NewCallNode() *CallNode {
	n := &CallNode{
		tracer:   packet.NewTracer(),
		inPort:   port.NewIn(),
		outPorts: []*port.OutPort{port.NewOut(), port.NewOut()},
		errPort:  port.NewOut(),
	}

	n.inPort.Accept(port.ListenFunc(n.forward))
	n.outPorts[0].Accept(port.ListenFunc(n.backward0))
	n.outPorts[1].Accept(port.ListenFunc(n.backward1))
	n.errPort.Accept(port.ListenFunc(n.catch))

	return n
}

// In returns the input port with the specified name.
func (n *CallNode) In(name string) *port.InPort {
	switch name {
	case node.PortIn:
		return n.inPort
	default:
	}
	return nil
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

	for {
		inPck, ok := <-inReader.Read()
		if !ok {
			return
		}
		n.tracer.Read(inReader, inPck)

		n.tracer.Sniff(inPck, packet.HandlerFunc(func(backPck *packet.Packet) {
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

func (n *CallNode) backward0(proc *process.Process) {
	outWriter0 := n.outPorts[0].Open(proc)

	for {
		backPck, ok := <-outWriter0.Receive()
		if !ok {
			return
		}

		n.tracer.Receive(outWriter0, backPck)
	}
}

func (n *CallNode) backward1(proc *process.Process) {
	outWriter1 := n.outPorts[1].Open(proc)

	for {
		backPck, ok := <-outWriter1.Receive()
		if !ok {
			return
		}

		n.tracer.Receive(outWriter1, backPck)
	}
}

func (n *CallNode) catch(proc *process.Process) {
	errWriter := n.errPort.Open(proc)

	for {
		backPck, ok := <-errWriter.Receive()
		if !ok {
			return
		}

		n.tracer.Receive(errWriter, backPck)
	}
}

// NewCallNodeCodec creates a new codec for CallNodeSpec.
func NewCallNodeCodec() scheme.Codec {
	return scheme.CodecWithType(func(spec *CallNodeSpec) (node.Node, error) {
		n := NewCallNode()

		return n, nil
	})
}
