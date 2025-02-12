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

// PipeNodeSpec holds the specification for creating a PipeNode.
type PipeNodeSpec struct {
	spec.Meta `json:",inline"`
}

// PipeNode processes an input packet and sends the result to multiple output ports.
type PipeNode struct {
	tracer   *packet.Tracer
	inPort   *port.InPort
	outPorts [2]*port.OutPort
	errPort  *port.OutPort
}

const KindPipe = "pipe"

var _ node.Node = (*PipeNode)(nil)

// NewPipeNodeCodec creates a new codec for PipeNodeSpec.
func NewPipeNodeCodec() scheme.Codec {
	return scheme.CodecWithType(func(spec *PipeNodeSpec) (node.Node, error) {
		return NewPipeNode(), nil
	})
}

// NewPipeNode creates a new PipeNode.
func NewPipeNode() *PipeNode {
	n := &PipeNode{
		tracer:   packet.NewTracer(),
		inPort:   port.NewIn(),
		outPorts: [2]*port.OutPort{port.NewOut(), port.NewOut()},
		errPort:  port.NewOut(),
	}

	n.inPort.AddListener(port.ListenFunc(n.forward))
	n.outPorts[0].AddListener(port.ListenFunc(n.backward0))
	n.outPorts[1].AddListener(port.ListenFunc(n.backward1))
	n.errPort.AddListener(port.ListenFunc(n.catch))

	return n
}

// In returns the input port with the specified name.
func (n *PipeNode) In(name string) *port.InPort {
	switch name {
	case node.PortIn:
		return n.inPort
	default:
		return nil
	}
}

// Out returns the output port with the specified name.
func (n *PipeNode) Out(name string) *port.OutPort {
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
	}
	return nil
}

// Close closes all ports associated with the node.
func (n *PipeNode) Close() error {
	n.inPort.Close()
	for _, outPort := range n.outPorts {
		outPort.Close()
	}
	n.errPort.Close()
	n.tracer.Close()
	return nil
}

func (n *PipeNode) forward(proc *process.Process) {
	inReader := n.inPort.Open(proc)
	var outWriter0 *packet.Writer

	for inPck := range inReader.Read() {
		n.tracer.Read(inReader, inPck)

		if outWriter0 == nil {
			outWriter0 = n.outPorts[0].Open(proc)
		}
		n.tracer.Write(outWriter0, inPck)
	}
}

func (n *PipeNode) backward0(proc *process.Process) {
	outWriter0 := n.outPorts[0].Open(proc)
	var outWriter1 *packet.Writer
	var errWriter *packet.Writer

	for backPck := range outWriter0.Receive() {
		outPcks := n.tracer.Writes(outWriter0)
		if len(outPcks) > 0 {
			n.tracer.Link(outPcks[0], backPck)
		}

		if _, ok := backPck.Payload().(types.Error); ok {
			if errWriter == nil {
				errWriter = n.errPort.Open(proc)
			}
			n.tracer.Write(errWriter, backPck)
		} else {
			if outWriter1 == nil {
				outWriter1 = n.outPorts[1].Open(proc)
			}
			n.tracer.Write(outWriter1, backPck)
		}

		n.tracer.Receive(outWriter0, nil)
	}
}

func (n *PipeNode) backward1(proc *process.Process) {
	outWriter := n.outPorts[1].Open(proc)

	for backPck := range outWriter.Receive() {
		n.tracer.Receive(outWriter, backPck)
	}
}

func (n *PipeNode) catch(proc *process.Process) {
	errWriter := n.errPort.Open(proc)

	for backPck := range errWriter.Receive() {
		n.tracer.Receive(errWriter, backPck)
	}
}
