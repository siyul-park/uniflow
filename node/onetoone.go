package node

import (
	"github.com/siyul-park/uniflow/packet"
	"github.com/siyul-park/uniflow/port"
	"github.com/siyul-park/uniflow/process"
)

// OneToOneNode represents a node that processes a packet from one input port and sends it to one output port.
type OneToOneNode struct {
	action  func(*process.Process, *packet.Packet) (*packet.Packet, *packet.Packet)
	tracer  *packet.Tracer
	inPort  *port.InPort
	outPort *port.OutPort
	errPort *port.OutPort
}

var _ Node = (*OneToOneNode)(nil)

// NewOneToOneNode creates a OneToOneNode with the specified action function.
func NewOneToOneNode(action func(*process.Process, *packet.Packet) (*packet.Packet, *packet.Packet)) *OneToOneNode {
	n := &OneToOneNode{
		action:  action,
		tracer:  packet.NewTracer(),
		inPort:  port.NewIn(),
		outPort: port.NewOut(),
		errPort: port.NewOut(),
	}

	if n.action != nil {
		n.inPort.AddListener(port.ListenFunc(n.forward))
		n.outPort.AddListener(port.ListenFunc(n.backward))
		n.errPort.AddListener(port.ListenFunc(n.catch))
	}

	return n
}

// In returns the input port with the specified name.
func (n *OneToOneNode) In(name string) *port.InPort {
	switch name {
	case PortIn:
		return n.inPort
	default:
		return nil
	}
}

// Out returns the output port for the specified name.
func (n *OneToOneNode) Out(name string) *port.OutPort {
	switch name {
	case PortOut:
		return n.outPort
	case PortError:
		return n.errPort
	default:
		return nil
	}
}

// Close closes all ports and releases resources.
func (n *OneToOneNode) Close() error {
	n.inPort.Close()
	n.outPort.Close()
	n.errPort.Close()
	n.tracer.Close()
	return nil
}

func (n *OneToOneNode) forward(proc *process.Process) {
	inReader := n.inPort.Open(proc)
	var outWriter *packet.Writer
	var errWriter *packet.Writer

	for inPck := range inReader.Read() {
		n.tracer.Read(inReader, inPck)

		if outPck, errPck := n.action(proc, inPck); errPck != nil {
			if errWriter == nil {
				errWriter = n.errPort.Open(proc)
			}
			n.tracer.Link(inPck, errPck)
			n.tracer.Write(errWriter, errPck)
		} else {
			if outWriter == nil {
				outWriter = n.outPort.Open(proc)
			}
			n.tracer.Link(inPck, outPck)
			n.tracer.Write(outWriter, outPck)
		}
	}
}

func (n *OneToOneNode) backward(proc *process.Process) {
	outWriter := n.outPort.Open(proc)

	for backPck := range outWriter.Receive() {
		n.tracer.Receive(outWriter, backPck)
	}
}

func (n *OneToOneNode) catch(proc *process.Process) {
	errWriter := n.errPort.Open(proc)

	for backPck := range errWriter.Receive() {
		n.tracer.Receive(errWriter, backPck)
	}
}
