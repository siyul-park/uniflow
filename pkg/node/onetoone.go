package node

import (
	"sync"

	"github.com/siyul-park/uniflow/pkg/packet"
	"github.com/siyul-park/uniflow/pkg/port"
	"github.com/siyul-park/uniflow/pkg/process"
)

// OneToOneNode represents a node that processes *packet.Packet with one input and one output.
type OneToOneNode struct {
	action  func(*process.Process, *packet.Packet) (*packet.Packet, *packet.Packet)
	ioPort  *port.Port
	inPort  *port.Port
	outPort *port.Port
	errPort *port.Port
	mu      sync.RWMutex
}

var _ Node = (*OneToOneNode)(nil)

// NewOneToOneNode creates a new OneToOneNode.
func NewOneToOneNode(action func(*process.Process, *packet.Packet) (*packet.Packet, *packet.Packet)) *OneToOneNode {
	n := &OneToOneNode{
		action:  action,
		ioPort:  port.New(),
		inPort:  port.New(),
		outPort: port.New(),
		errPort: port.New(),
	}

	if n.action != nil {
		n.ioPort.AddInitHook(port.InitHookFunc(func(proc *process.Process) {
			n.mu.RLock()
			defer n.mu.RUnlock()

			ioStream := n.ioPort.Open(proc)

			n.forward(proc, ioStream, ioStream)
		}))
		n.inPort.AddInitHook(port.InitHookFunc(func(proc *process.Process) {
			n.mu.RLock()
			defer n.mu.RUnlock()

			inStream := n.inPort.Open(proc)
			outStream := n.outPort.Open(proc)

			n.forward(proc, inStream, outStream)
		}))
		n.outPort.AddInitHook(port.InitHookFunc(func(proc *process.Process) {
			n.mu.RLock()
			defer n.mu.RUnlock()

			outStream := n.outPort.Open(proc)

			n.backward(proc, outStream)
		}))
		n.errPort.AddInitHook(port.InitHookFunc(func(proc *process.Process) {
			n.mu.RLock()
			defer n.mu.RUnlock()

			errStream := n.errPort.Open(proc)

			n.backward(proc, errStream)
		}))
	}

	return n
}

// Port returns the specified port.
func (n *OneToOneNode) Port(name string) (*port.Port, bool) {
	n.mu.RLock()
	defer n.mu.RUnlock()

	switch name {
	case PortIO:
		return n.ioPort, true
	case PortIn:
		return n.inPort, true
	case PortOut:
		return n.outPort, true
	case PortErr:
		return n.errPort, true
	default:
	}

	return nil, false
}

// Close closes all.
func (n *OneToOneNode) Close() error {
	n.mu.Lock()
	defer n.mu.Unlock()

	n.ioPort.Close()
	n.inPort.Close()
	n.outPort.Close()
	n.errPort.Close()

	return nil
}

func (n *OneToOneNode) forward(proc *process.Process, inStream, outStream *port.Stream) {
	n.mu.RLock()
	defer n.mu.RUnlock()

	errStream := n.errPort.Open(proc)

	for {
		inPck, ok := <-inStream.Receive()
		if !ok {
			return
		}

		if outPck, errPck := n.action(proc, inPck); errPck != nil {
			if errPck == inPck {
				errPck = packet.New(errPck.Payload())
			}
			proc.Stack().Link(inPck.ID(), errPck.ID())
			if errStream.Links() > 0 {
				proc.Stack().Push(errPck.ID(), inStream.ID())
				errStream.Send(errPck)
			} else {
				inStream.Send(errPck)
			}
		} else if outPck != nil && outStream.Links() > 0 {
			if outPck == inPck {
				outPck = packet.New(outPck.Payload())
			}
			proc.Stack().Link(inPck.ID(), outPck.ID())
			if outStream != inStream {
				proc.Stack().Push(outPck.ID(), inStream.ID())
				outStream.Send(outPck)
			} else {
				inStream.Send(outPck)
			}
		} else {
			proc.Stack().Clear(inPck.ID())
		}
	}
}

func (n *OneToOneNode) backward(proc *process.Process, outStream *port.Stream) {
	n.mu.RLock()
	defer n.mu.RUnlock()

	var ioStream *port.Stream
	var inStream *port.Stream

	for {
		backPck, ok := <-outStream.Receive()
		if !ok {
			return
		}

		if ioStream == nil {
			ioStream = n.ioPort.Open(proc)
		}
		if inStream == nil {
			inStream = n.inPort.Open(proc)
		}

		if _, ok := proc.Stack().Pop(backPck.ID(), ioStream.ID()); ok {
			ioStream.Send(backPck)
		} else if _, ok := proc.Stack().Pop(backPck.ID(), inStream.ID()); ok {
			inStream.Send(backPck)
		}
	}
}
