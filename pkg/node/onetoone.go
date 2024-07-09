package node

import (
	"sync"

	"github.com/siyul-park/uniflow/pkg/packet"
	"github.com/siyul-park/uniflow/pkg/port"
	"github.com/siyul-park/uniflow/pkg/process"
)

// OneToOneNode represents a node that processes a packet from one input port and sends it to one output port.
type OneToOneNode struct {
	action  func(*process.Process, *packet.Packet) (*packet.Packet, *packet.Packet)
	tracer  *packet.Tracer
	inPort  *port.InPort
	outPort *port.OutPort
	errPort *port.OutPort
	mu      sync.RWMutex
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
		n.inPort.Accept(port.ListenFunc(n.forward))
		n.outPort.Accept(port.ListenFunc(n.backward))
		n.errPort.Accept(port.ListenFunc(n.catch))
	}

	return n
}

// In returns the input port with the specified name.
func (n *OneToOneNode) In(name string) *port.InPort {
	n.mu.RLock()
	defer n.mu.RUnlock()

	switch name {
	case PortIn:
		return n.inPort
	default:
		return nil
	}
}

// Out returns the output port for the specified name.
func (n *OneToOneNode) Out(name string) *port.OutPort {
	n.mu.RLock()
	defer n.mu.RUnlock()

	switch name {
	case PortOut:
		return n.outPort
	case PortErr:
		return n.errPort
	default:
		return nil
	}
}

// Close closes all ports and releases resources.
func (n *OneToOneNode) Close() error {
	n.mu.Lock()
	defer n.mu.Unlock()

	n.inPort.Close()
	n.outPort.Close()
	n.errPort.Close()
	n.tracer.Close()

	return nil
}

func (n *OneToOneNode) forward(proc *process.Process) {
	n.mu.RLock()
	defer n.mu.RUnlock()

	inReader := n.inPort.Open(proc)
	outWriter := n.outPort.Open(proc)
	errWriter := n.errPort.Open(proc)

	for {
		inPck, ok := <-inReader.Read()
		if !ok {
			return
		}
		n.tracer.Read(inReader, inPck)

		if outPck, errPck := n.action(proc, inPck); errPck != nil {
			n.tracer.Transform(inPck, errPck)
			n.tracer.Write(errWriter, errPck)
		} else {
			n.tracer.Transform(inPck, outPck)
			n.tracer.Write(outWriter, outPck)
		}
	}
}

func (n *OneToOneNode) backward(proc *process.Process) {
	n.mu.RLock()
	defer n.mu.RUnlock()

	outWriter := n.outPort.Open(proc)

	for {
		backPck, ok := <-outWriter.Receive()
		if !ok {
			return
		}

		n.tracer.Receive(outWriter, backPck)
	}
}

func (n *OneToOneNode) catch(proc *process.Process) {
	n.mu.RLock()
	defer n.mu.RUnlock()

	errWriter := n.errPort.Open(proc)

	for {
		backPck, ok := <-errWriter.Receive()
		if !ok {
			return
		}

		n.tracer.Receive(errWriter, backPck)
	}
}
