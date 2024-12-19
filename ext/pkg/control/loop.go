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

// LoopNodeSpec defines the specifications for creating a LoopNode.
type LoopNodeSpec struct {
	spec.Meta `map:",inline"`
}

// LoopNode processes input data in batches, splitting packets into sub-packets and handling them accordingly.
type LoopNode struct {
	tracer   *packet.Tracer
	inPort   *port.InPort
	outPorts [2]*port.OutPort
	errPort  *port.OutPort
}

const KindLoop = "loop"

var _ node.Node = (*LoopNode)(nil)

// NewLoopNodeCodec creates a new codec for LoopNodeSpec.
func NewLoopNodeCodec() scheme.Codec {
	return scheme.CodecWithType(func(spec *LoopNodeSpec) (node.Node, error) {
		return NewLoopNode(), nil
	})
}

// NewLoopNode creates a new LoopNode with default configurations.
func NewLoopNode() *LoopNode {
	n := &LoopNode{
		tracer:   packet.NewTracer(),
		inPort:   port.NewIn(),
		outPorts: [2]*port.OutPort{port.NewOut(), port.NewOut()},
		errPort:  port.NewOut(),
	}

	n.inPort.AddListener(port.ListenFunc(n.forward))
	n.outPorts[0].AddListener(n.backward(0))
	n.outPorts[1].AddListener(n.backward(1))
	n.errPort.AddListener(port.ListenFunc(n.catch))

	return n
}

// In returns the input port with the specified name.
func (n *LoopNode) In(name string) *port.InPort {
	switch name {
	case node.PortIn:
		return n.inPort
	default:
		return nil
	}
}

// Out returns the output port with the specified name.
func (n *LoopNode) Out(name string) *port.OutPort {
	switch name {
	case node.PortOut:
		return n.outPorts[0]
	case node.PortError:
		return n.errPort
	default:
		if node.NameOfPort(name) == node.PortOut {
			index, ok := node.IndexOfPort(name)
			if ok && index < len(n.outPorts) {
				return n.outPorts[index]
			}
		}
		return nil
	}
}

// Close closes all ports associated with the node.
func (n *LoopNode) Close() error {
	n.inPort.Close()
	for _, outPort := range n.outPorts {
		outPort.Close()
	}
	n.errPort.Close()
	n.tracer.Close()
	return nil
}

func (n *LoopNode) forward(proc *process.Process) {
	inReader := n.inPort.Open(proc)
	outWriter0 := n.outPorts[0].Open(proc)
	outWriter1 := n.outPorts[1].Open(proc)
	errWriter := n.errPort.Open(proc)

	for inPck := range inReader.Read() {
		n.tracer.Read(inReader, inPck)

		inPayload := inPck.Payload()

		var outPayloads []types.Value
		if v, ok := inPayload.(types.Slice); ok {
			outPayloads = v.Values()
		} else {
			outPayloads = []types.Value{inPayload}
		}

		outPcks := make([]*packet.Packet, len(outPayloads))
		for i, outPayload := range outPayloads {
			outPck := packet.New(outPayload)
			outPcks[i] = outPck
			n.tracer.Transform(inPck, outPck)
		}

		n.tracer.AddHook(inPck, packet.HookFunc(func(backPck *packet.Packet) {
			n.tracer.Transform(inPck, backPck)
			if _, ok := backPck.Payload().(types.Error); ok {
				n.tracer.Write(errWriter, backPck)
			} else {
				n.tracer.Write(outWriter1, backPck)
			}
		}))

		for _, outPck := range outPcks {
			n.tracer.Write(outWriter0, outPck)
		}
	}
}

func (n *LoopNode) backward(index int) port.Listener {
	return port.ListenFunc(func(proc *process.Process) {
		outWriter := n.outPorts[index].Open(proc)

		for backPck := range outWriter.Receive() {
			n.tracer.Receive(outWriter, backPck)
		}
	})
}

func (n *LoopNode) catch(proc *process.Process) {
	errWriter := n.errPort.Open(proc)

	for backPck := range errWriter.Receive() {
		n.tracer.Receive(errWriter, backPck)
	}
}
