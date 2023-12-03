package node

import (
	"sync"

	"github.com/oklog/ulid/v2"
	"github.com/siyul-park/uniflow/pkg/packet"
	"github.com/siyul-park/uniflow/pkg/port"
	"github.com/siyul-park/uniflow/pkg/process"
)

// OneToOneNodeConfig is a config for ActionNode.
type OneToOneNodeConfig struct {
	ID     ulid.ULID
	Action func(*process.Process, *packet.Packet) (*packet.Packet, *packet.Packet)
}

// OneToOneNode provide process *packet.Packet one source and onde distance.
type OneToOneNode struct {
	id      ulid.ULID
	action  func(*process.Process, *packet.Packet) (*packet.Packet, *packet.Packet)
	ioPort  *port.Port
	inPort  *port.Port
	outPort *port.Port
	errPort *port.Port
	mu      sync.RWMutex
}

var _ Node = (*OneToOneNode)(nil)

// NewOneToOneNode returns a new OneToOneNode.
func NewOneToOneNode(config OneToOneNodeConfig) *OneToOneNode {
	id := config.ID
	action := config.Action

	if id == (ulid.ULID{}) {
		id = ulid.Make()
	}
	if action == nil {
		action = func(_ *process.Process, _ *packet.Packet) (*packet.Packet, *packet.Packet) {
			return nil, nil
		}
	}

	n := &OneToOneNode{
		id:      id,
		action:  action,
		ioPort:  port.New(),
		inPort:  port.New(),
		outPort: port.New(),
		errPort: port.New(),
	}

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

	return n
}

func (n *OneToOneNode) ID() ulid.ULID {
	n.mu.RLock()
	defer n.mu.RUnlock()

	return n.id
}

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

func (n *OneToOneNode) Close() error {
	n.mu.Lock()
	defer n.mu.Unlock()

	n.ioPort.Close()
	n.inPort.Close()
	n.outPort.Close()
	n.errPort.Close()

	return nil
}

func (n *OneToOneNode) forward(proc *process.Process, inStream *port.Stream, outStream *port.Stream) {
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
