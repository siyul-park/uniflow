package node

import (
	"github.com/siyul-park/uniflow/pkg/node"
	"github.com/siyul-park/uniflow/pkg/packet"
	"github.com/siyul-park/uniflow/pkg/port"
	"github.com/siyul-park/uniflow/pkg/process"
	"github.com/siyul-park/uniflow/pkg/scheme"
	"github.com/siyul-park/uniflow/pkg/spec"
	"github.com/siyul-park/uniflow/pkg/types"
)

// TryNodeSpec defines the specification for creating a TryNode.
type TryNodeSpec struct {
	spec.Meta `json:",inline"`
}

// TryNode represents a node that processes packets and handles errors.
type TryNode struct {
	tracer  *packet.Tracer
	inPort  *port.InPort
	outPort *port.OutPort
	errPort *port.OutPort
}

const KindTry = "try"

var _ node.Node = (*TryNode)(nil)

// NewTryNodeCodec creates a codec for decoding TryNodeSpec.
func NewTryNodeCodec() scheme.Codec {
	return scheme.CodecWithType(func(_ *TryNodeSpec) (node.Node, error) {
		return NewTryNode(), nil
	})
}

// NewTryNode creates a new TryNode.
func NewTryNode() *TryNode {
	n := &TryNode{
		tracer:  packet.NewTracer(),
		inPort:  port.NewIn(),
		outPort: port.NewOut(),
		errPort: port.NewOut(),
	}

	n.inPort.AddListener(port.ListenFunc(n.forward))
	n.outPort.AddListener(port.ListenFunc(n.backward))
	n.errPort.AddListener(port.ListenFunc(n.catch))

	return n
}

// In returns the input port for the given name.
func (n *TryNode) In(name string) *port.InPort {
	switch name {
	case node.PortIn:
		return n.inPort
	default:
		return nil
	}
}

// Out returns the output port for the given name.
func (n *TryNode) Out(name string) *port.OutPort {
	switch name {
	case node.PortOut:
		return n.outPort
	case node.PortError:
		return n.errPort
	default:
		return nil
	}
}

// Close closes the TryNode and its ports.
func (n *TryNode) Close() error {
	n.inPort.Close()
	n.outPort.Close()
	n.errPort.Close()
	n.tracer.Close()
	return nil
}

func (n *TryNode) forward(proc *process.Process) {
	inReader := n.inPort.Open(proc)
	var outWriter *packet.Writer

	for inPck := range inReader.Read() {
		n.tracer.Read(inReader, inPck)

		if outWriter == nil {
			outWriter = n.outPort.Open(proc)
		}
		n.tracer.Write(outWriter, inPck)
	}
}

func (n *TryNode) backward(proc *process.Process) {
	outWriter := n.outPort.Open(proc)
	var errWriter *packet.Writer

	for backPck := range outWriter.Receive() {
		if _, ok := backPck.Payload().(types.Error); ok {
			outPcks := n.tracer.Writes(outWriter)
			if len(outPcks) > 0 {
				n.tracer.Link(outPcks[0], backPck)
			}

			if errWriter == nil {
				errWriter = n.errPort.Open(proc)
			}
			n.tracer.Write(errWriter, backPck)
			n.tracer.Receive(outWriter, nil)
		} else {
			n.tracer.Receive(outWriter, backPck)
		}
	}
}

func (n *TryNode) catch(proc *process.Process) {
	errWriter := n.errPort.Open(proc)

	for backPck := range errWriter.Receive() {
		n.tracer.Receive(errWriter, backPck)
	}
}
