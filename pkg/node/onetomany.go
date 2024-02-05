package node

import (
	"sync"

	"github.com/siyul-park/uniflow/pkg/packet"
	"github.com/siyul-park/uniflow/pkg/port"
	"github.com/siyul-park/uniflow/pkg/process"
)

// OneToManyNode represents a node that processes *packet.Packet with one input and many outputs.
type OneToManyNode struct {
	action   func(*process.Process, *packet.Packet) ([]*packet.Packet, *packet.Packet)
	inPort   *port.Port
	outPorts []*port.Port
	errPort  *port.Port
	mu       sync.RWMutex
}

var _ Node = (*OneToManyNode)(nil)

// NewOneToManyNode creates a new OneToManyNode.
func NewOneToManyNode(action func(*process.Process, *packet.Packet) ([]*packet.Packet, *packet.Packet)) *OneToManyNode {
	n := &OneToManyNode{
		action:   action,
		inPort:   port.New(),
		outPorts: nil,
		errPort:  port.New(),
	}

	if n.action != nil {
		n.inPort.AddInitHook(port.InitHookFunc(n.forward))
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
func (n *OneToManyNode) Port(name string) *port.Port {
	n.mu.Lock()
	defer n.mu.Unlock()

	switch name {
	case PortIn:
		return n.inPort
	case PortErr:
		return n.errPort
	default:
		if i, ok := IndexOfMultiPort(PortOut, name); ok {
			for j := 0; j <= i; j++ {
				if len(n.outPorts) <= j {
					outPort := port.New()
					if n.action != nil {
						outPort.AddInitHook(port.InitHookFunc(func(proc *process.Process) {
							n.mu.RLock()
							defer n.mu.RUnlock()

							outStream := outPort.Open(proc)

							n.backward(proc, outStream)
						}))
					}
					n.outPorts = append(n.outPorts, outPort)
				}
			}

			return n.outPorts[i]
		}
	}

	return nil
}

// Close closes all ports.
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
	inStream := n.inPort.Open(proc)
	outStreams := make([]*port.Stream, len(n.outPorts))
	for i, p := range n.outPorts {
		outStreams[i] = p.Open(proc)
	}
	errStream := n.errPort.Open(proc)

	for _, outStream := range outStreams {
		outStream := outStream
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

		if outPcks, errPck := n.action(proc, inPck); errPck != nil {
			proc.Graph().Add(inPck.ID(), errPck.ID())
			if errStream.Links() > 0 {
				proc.Stack().Push(errPck.ID(), inStream.ID())
				errStream.Send(errPck)
			} else {
				inStream.Send(errPck)
			}
		} else if len(outPcks) > 0 && len(outStreams) > 0 {
			for i, outPck := range outPcks {
				if len(outStreams) <= i || outPck == nil {
					continue
				}
				outStream := outStreams[i]
				if outStream.Links() == 0 {
					continue
				}

				proc.Graph().Add(inPck.ID(), outPck.ID())
				proc.Stack().Push(outPck.ID(), inStream.ID())
				outStream.Send(outPck)
			}
		} else {
			proc.Stack().Clear(inPck.ID())
		}
	}
}

func (n *OneToManyNode) backward(proc *process.Process, outStream *port.Stream) {
	var inStream *port.Stream

	for {
		backPck, ok := <-outStream.Receive()
		if !ok {
			return
		}

		if _, ok := proc.Stack().Pop(backPck.ID(), outStream.ID()); !ok {
			continue
		}

		if inStream == nil {
			inStream = n.inPort.Open(proc)
		}

		if _, ok := proc.Stack().Pop(backPck.ID(), inStream.ID()); ok {
			inStream.Send(backPck)
		}
	}
}
