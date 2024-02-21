package node

import (
	"sync"

	"github.com/samber/lo"
	"github.com/siyul-park/uniflow/pkg/packet"
	"github.com/siyul-park/uniflow/pkg/port"
	"github.com/siyul-park/uniflow/pkg/process"
)

// ManyToOneNode represents a node that processes *packet.Packet with many inputs and one output.
type ManyToOneNode struct {
	action      func(*process.Process, []*packet.Packet) (*packet.Packet, *packet.Packet)
	inPorts     []*port.Port
	outPort     *port.Port
	errPort     *port.Port
	forwardOnce *port.InitOnceHook
	mu          sync.RWMutex
}

var _ Node = (*ManyToOneNode)(nil)

// NewManyToOneNode creates a new ManyToOneNode.
func NewManyToOneNode(action func(*process.Process, []*packet.Packet) (*packet.Packet, *packet.Packet)) *ManyToOneNode {
	n := &ManyToOneNode{
		action:  action,
		outPort: port.New(),
		errPort: port.New(),
	}
	n.forwardOnce = &port.InitOnceHook{
		Hook: port.InitHookFunc(n.forward),
	}

	if n.action != nil {
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
func (n *ManyToOneNode) Port(name string) *port.Port {
	n.mu.Lock()
	defer n.mu.Unlock()

	switch name {
	case PortOut:
		return n.outPort
	case PortErr:
		return n.errPort
	default:
		if i, ok := IndexOfMultiPort(PortIn, name); ok {
			for j := 0; j <= i; j++ {
				if len(n.inPorts) <= j {
					inPort := port.New()
					if n.action != nil {
						inPort.AddInitHook(n.forwardOnce)
					}
					n.inPorts = append(n.inPorts, inPort)
				}
			}

			return n.inPorts[i]
		}
	}

	return nil
}

// Close closes all ports.
func (n *ManyToOneNode) Close() error {
	n.mu.Lock()
	defer n.mu.Unlock()

	for _, p := range n.inPorts {
		p.Close()
	}
	n.outPort.Close()
	n.errPort.Close()

	return nil
}

func (n *ManyToOneNode) forward(proc *process.Process) {
	var inStreams []*port.Stream
	for _, p := range n.inPorts {
		if p.Links() > 0 {
			inStreams = append(inStreams, p.Open(proc))
		}
	}
	outStream := n.outPort.Open(proc)
	errStream := n.errPort.Open(proc)

	outStream.AddSendHook(port.SendHookFunc(func(pck *packet.Packet) {
		proc.Stack().Push(pck.ID(), outStream.ID())
	}))
	errStream.AddSendHook(port.SendHookFunc(func(pck *packet.Packet) {
		proc.Stack().Push(pck.ID(), errStream.ID())
	}))

	buffers := make([][]*packet.Packet, len(inStreams))

	consume := func(index int) {
		inPcks := make([]*packet.Packet, len(inStreams))
		for i, buffer := range buffers {
			if len(buffer) > index {
				inPcks[i] = buffer[index]
				buffers[i] = append(buffer[:index], buffer[index+1:]...)
			}
		}

		forward := func(outStream *port.Stream, outPck *packet.Packet, backward bool) {
			inStreams = lo.Filter[*port.Stream](inStreams, func(_ *port.Stream, i int) bool {
				return inPcks[i] != nil
			})
			inPcks = lo.Filter[*packet.Packet](inPcks, func(item *packet.Packet, _ int) bool {
				return item != nil
			})

			for _, inPck := range inPcks {
				proc.Graph().Add(inPck.ID(), outPck.ID())
			}

			if outStream.Links() > 0 {
				for _, inStream := range inStreams {
					proc.Stack().Push(outPck.ID(), inStream.ID())
				}
				outStream.Send(outPck)
			} else if backward {
				for _, inStream := range inStreams {
					inStream.Send(outPck)
				}
			} else {
				proc.Stack().Clear(outPck.ID())
			}
		}

		if outPck, errPck := n.action(proc, inPcks); errPck != nil {
			forward(errStream, errPck, true)
		} else if outPck != nil {
			forward(outStream, outPck, false)
		} else {
			if lo.Count[*packet.Packet](inPcks, nil) == 0 {
				for _, inPck := range inPcks {
					proc.Stack().Clear(inPck.ID())
				}
			} else {
				for i, inPck := range inPcks {
					if inPck != nil {
						buffers[i] = append(buffers[i][:index+1], buffers[i][index:]...)
						buffers[i][index] = inPck
					}
				}
			}
		}
	}

	mu := &sync.Mutex{}
	wg := &sync.WaitGroup{}
	for i, inStream := range inStreams {
		wg.Add(1)
		go func(i int, inStream *port.Stream) {
			defer wg.Done()
			for {
				inPck, ok := <-inStream.Receive()
				if !ok {
					return
				}

				mu.Lock()

				buffers[i] = append(buffers[i], inPck)
				consume(len(buffers[i]) - 1)

				mu.Unlock()
			}
		}(i, inStream)
	}
	wg.Wait()
}

func (n *ManyToOneNode) backward(proc *process.Process, outStream *port.Stream) {
	var inStreams []*port.Stream

	for {
		backPck, ok := <-outStream.Receive()
		if !ok {
			return
		}

		if _, ok := proc.Stack().Pop(backPck.ID(), outStream.ID()); !ok {
			continue
		}

		if inStreams == nil {
			inStreams = make([]*port.Stream, len(n.inPorts))
			for i, p := range n.inPorts {
				inStreams[i] = p.Open(proc)
			}
		}

		for i := len(inStreams) - 1; i >= 0; i-- {
			inStream := inStreams[i]
			if _, ok := proc.Stack().Pop(backPck.ID(), inStream.ID()); ok {
				inStream.Send(backPck)
			}
		}
	}
}
