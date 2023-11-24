package node

import (
	"sync"

	"github.com/oklog/ulid/v2"
	"github.com/siyul-park/uniflow/pkg/packet"
	"github.com/siyul-park/uniflow/pkg/port"
	"github.com/siyul-park/uniflow/pkg/process"
)

type (
	// OneToManyNodeConfig is a config for ActionNode.
	OneToManyNodeConfig struct {
		ID     ulid.ULID
		Action func(*process.Process, *packet.Packet) ([]*packet.Packet, *packet.Packet)
	}

	// OneToManyNode provide process *packet.Packet one source and many distance.
	OneToManyNode struct {
		id       ulid.ULID
		action   func(*process.Process, *packet.Packet) ([]*packet.Packet, *packet.Packet)
		inPort   *port.Port
		outPorts []*port.Port
		errPort  *port.Port
		mu       sync.RWMutex
	}
)

var _ Node = &OneToManyNode{}

// NewOneToManyNode returns a new OneToManyNode.
func NewOneToManyNode(config OneToManyNodeConfig) *OneToManyNode {
	id := config.ID
	action := config.Action

	if id == (ulid.ULID{}) {
		id = ulid.Make()
	}
	if action == nil {
		action = func(_ *process.Process, _ *packet.Packet) ([]*packet.Packet, *packet.Packet) {
			return nil, nil
		}
	}

	n := &OneToManyNode{
		id:       id,
		action:   action,
		inPort:   port.New(),
		outPorts: nil,
		errPort:  port.New(),
	}

	n.inPort.AddInitHook(port.InitHookFunc(n.forward))
	n.errPort.AddInitHook(port.InitHookFunc(func(proc *process.Process) {
		n.mu.RLock()
		defer n.mu.RUnlock()

		errStream := n.errPort.Open(proc)

		n.backward(proc, errStream)
	}))

	return n
}

func (n *OneToManyNode) ID() ulid.ULID {
	n.mu.RLock()
	defer n.mu.RUnlock()

	return n.id
}

func (n *OneToManyNode) Port(name string) (*port.Port, bool) {
	n.mu.Lock()
	defer n.mu.Unlock()

	switch name {
	case PortIn:
		return n.inPort, true
	case PortErr:
		return n.errPort, true
	default:
	}

	if i, ok := port.GetIndex(PortOut, name); ok {
		for j := 0; j <= i; j++ {
			if len(n.outPorts) <= j {
				outPort := port.New()
				outPort.AddInitHook(port.InitHookFunc(func(proc *process.Process) {
					n.mu.RLock()
					defer n.mu.RUnlock()

					outStream := outPort.Open(proc)

					n.backward(proc, outStream)
				}))
				n.outPorts = append(n.outPorts, outPort)
			}
		}

		return n.outPorts[i], true
	}

	return nil, false
}

func (n *OneToManyNode) Close() error {
	n.mu.Lock()
	defer n.mu.Unlock()

	n.inPort.Close()
	for _, p := range n.outPorts {
		p.Close()
	}
	n.errPort.Close()

	return nil
}

func (n *OneToManyNode) forward(proc *process.Process) {
	n.mu.RLock()
	defer n.mu.RUnlock()

	inStream := n.inPort.Open(proc)
	outStreams := make([]*port.Stream, len(n.outPorts))
	for i, p := range n.outPorts {
		outStreams[i] = p.Open(proc)
	}
	errStream := n.errPort.Open(proc)

	for func() bool {
		inPck, ok := <-inStream.Receive()
		if !ok {
			return false
		}

		if outPcks, errPck := n.action(proc, inPck); errPck != nil {
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
		} else if len(outPcks) > 0 && len(outStreams) > 0 {
			var ok bool
			for i, outPck := range outPcks {
				if len(outStreams) <= i {
					break
				}
				if outPck == nil {
					continue
				}
				outStream := outStreams[i]

				if outStream.Links() > 0 {
					if outPck == inPck {
						outPck = packet.New(outPck.Payload())
					}
					proc.Stack().Link(inPck.ID(), outPck.ID())
					proc.Stack().Push(outPck.ID(), inStream.ID())
					outStream.Send(outPck)

					ok = true
				}
			}

			if !ok {
				proc.Stack().Clear(inPck.ID())
			}
		} else {
			proc.Stack().Clear(inPck.ID())
		}

		return true
	}() {
	}
}

func (n *OneToManyNode) backward(proc *process.Process, outStream *port.Stream) {
	n.mu.RLock()
	defer n.mu.RUnlock()

	var inStream *port.Stream

	for {
		backPck, ok := <-outStream.Receive()
		if !ok {
			return
		}

		if inStream == nil {
			inStream = n.inPort.Open(proc)
		}

		if _, ok := proc.Stack().Pop(backPck.ID(), inStream.ID()); ok {
			inStream.Send(backPck)
		}
	}
}
