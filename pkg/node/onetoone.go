package node

import (
	"sync"

	"github.com/siyul-park/uniflow/pkg/packet"
	"github.com/siyul-park/uniflow/pkg/port"
	"github.com/siyul-park/uniflow/pkg/process"
)

// OneToOneNode represents a node with one input and one output port.
type OneToOneNode struct {
	action  func(*process.Process, *packet.Packet) (*packet.Packet, *packet.Packet)
	ioPort  *port.InPort
	inPort  *port.InPort
	outPort *port.OutPort
	errPort *port.OutPort
	mu      sync.RWMutex
}

var _ Node = (*OneToOneNode)(nil)

// NewOneToOneNode creates a new OneToOneNode instance with the given action function.
func NewOneToOneNode(action func(*process.Process, *packet.Packet) (*packet.Packet, *packet.Packet)) *OneToOneNode {
	n := &OneToOneNode{
		action:  action,
		ioPort:  port.NewIn(),
		inPort:  port.NewIn(),
		outPort: port.NewOut(),
		errPort: port.NewOut(),
	}

	if n.action != nil {
		n.ioPort.AddHandler(port.HandlerFunc(n.turn))
		n.inPort.AddHandler(port.HandlerFunc(n.forward))
		n.outPort.AddHandler(port.HandlerFunc(n.backward))
		n.errPort.AddHandler(port.HandlerFunc(n.catch))
	}

	return n
}

// In returns the input port with the specified name.
func (n *OneToOneNode) In(name string) *port.InPort {
	n.mu.RLock()
	defer n.mu.RUnlock()

	switch name {
	case PortIO:
		return n.ioPort
	case PortIn:
		return n.inPort
	default:
	}

	return nil
}

// Out returns the output port with the specified name.
func (n *OneToOneNode) Out(name string) *port.OutPort {
	n.mu.RLock()
	defer n.mu.RUnlock()

	switch name {
	case PortOut:
		return n.outPort
	case PortErr:
		return n.errPort
	default:
	}

	return nil
}

// Close closes all ports associated with the node.
func (n *OneToOneNode) Close() error {
	n.mu.Lock()
	defer n.mu.Unlock()

	n.ioPort.Close()
	n.inPort.Close()
	n.outPort.Close()
	n.errPort.Close()

	return nil
}

func (n *OneToOneNode) turn(proc *process.Process) {
	n.mu.RLock()
	defer n.mu.RUnlock()

	ioReader := n.ioPort.Open(proc)

	for {
		inPck, ok := <-ioReader.Read()
		if !ok {
			return
		}

		if outPck, errPck := n.action(proc, inPck); errPck != nil {
			proc.Stack().Add(inPck, errPck)
			n.throw(proc, errPck)
		} else if outPck != nil {
			proc.Stack().Add(inPck, outPck)
			ioReader.Receive(outPck)
		} else {
			proc.Stack().Clear(inPck)
		}
	}
}

func (n *OneToOneNode) forward(proc *process.Process) {
	n.mu.RLock()
	defer n.mu.RUnlock()

	inReader := n.inPort.Open(proc)
	outWriter := n.outPort.Open(proc)

	for {
		inPck, ok := <-inReader.Read()
		if !ok {
			return
		}

		if outPck, errPck := n.action(proc, inPck); errPck != nil {
			proc.Stack().Add(inPck, errPck)
			n.throw(proc, errPck)
		} else if outPck != nil {
			proc.Stack().Add(inPck, outPck)
			outWriter.Write(outPck)
		} else {
			proc.Stack().Clear(inPck)
		}
	}
}

func (n *OneToOneNode) backward(proc *process.Process) {
	n.mu.RLock()
	defer n.mu.RUnlock()

	inReader := n.inPort.Open(proc)
	outWriter := n.outPort.Open(proc)

	for {
		backPck, ok := <-outWriter.Receive()
		if !ok {
			return
		}

		inReader.Receive(backPck)
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

		n.receive(proc, backPck)
	}
}

func (n *OneToOneNode) throw(proc *process.Process, errPck *packet.Packet) {
	errWriter := n.errPort.Open(proc)

	if errWriter.Links() > 0 {
		errWriter.Write(errPck)
	} else {
		n.receive(proc, errPck)
	}
}

func (n *OneToOneNode) receive(proc *process.Process, backPck *packet.Packet) {
	ioReader := n.ioPort.Open(proc)
	inReader := n.inPort.Open(proc)

	ioCost := ioReader.Cost(backPck)
	inCost := inReader.Cost(backPck)

	if ioCost < inCost {
		ioReader.Receive(backPck)
	} else {
		inReader.Receive(backPck)
	}
}
