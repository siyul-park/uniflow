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
func (n *OneToOneNode) Port(name string) *port.Port {
	n.mu.RLock()
	defer n.mu.RUnlock()

	switch name {
	case PortIO:
		return n.ioPort
	case PortIn:
		return n.inPort
	case PortOut:
		return n.outPort
	case PortErr:
		return n.errPort
	default:
	}

	return nil
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
	errStream := n.errPort.Open(proc)

	if inStream != outStream {
		outStream.AddSendHook(port.SendHookFunc(func(pck *packet.Packet) {
			proc.Stack().Push(pck.ID(), outStream.ID())
		}))
	}
	errStream.AddSendHook(port.SendHookFunc(func(pck *packet.Packet) {
		proc.Stack().Push(pck.ID(), errStream.ID())
	}))

	for {
		inPck, ok := <-inStream.Receive()
		if !ok {
			return
		}

		forward := func(outStream *port.Stream, outPck *packet.Packet, backward bool) {
			proc.Graph().Add(inPck.ID(), outPck.ID())
			if outStream.Links() > 0 {
				if outStream != inStream {
					proc.Stack().Push(outPck.ID(), inStream.ID())
				}
				outStream.Send(outPck)
			} else if backward {
				inStream.Send(outPck)
			} else {
				proc.Stack().Clear(outPck.ID())
			}
		}

		if outPck, errPck := n.action(proc, inPck); errPck != nil {
			forward(errStream, errPck, true)
		} else if outPck != nil {
			forward(outStream, outPck, false)
		} else {
			proc.Stack().Clear(inPck.ID())
		}
	}
}

func (n *OneToOneNode) backward(proc *process.Process, outStream *port.Stream) {
	var ioStream *port.Stream
	var inStream *port.Stream

	for {
		backPck, ok := <-outStream.Receive()
		if !ok {
			return
		}

		if _, ok := proc.Stack().Pop(backPck.ID(), outStream.ID()); !ok {
			continue
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
