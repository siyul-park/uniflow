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
func (n *ManyToOneNode) Port(name string) (*port.Port, bool) {
	n.mu.Lock()
	defer n.mu.Unlock()

	switch name {
	case PortOut:
		return n.outPort, true
	case PortErr:
		return n.errPort, true
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

			return n.inPorts[i], true
		}
	}

	return nil, false
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
	n.mu.RLock()
	defer n.mu.RUnlock()

	var inStreams []*port.Stream
	for _, p := range n.inPorts {
		if p.Links() > 0 {
			inStreams = append(inStreams, p.Open(proc))
		}
	}
	outStream := n.outPort.Open(proc)
	errStream := n.errPort.Open(proc)

	buffers := make([][]*packet.Packet, len(inStreams))
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

				func() {
					mu.Lock()
					defer mu.Unlock()

					buffers[i] = append(buffers[i], inPck)
					bufferLen := len(buffers[i])

					inPcks := make([]*packet.Packet, len(inStreams))
					inPckLen := 0
					for i, buffer := range buffers {
						if len(buffer) >= bufferLen {
							inPcks[i] = buffer[bufferLen-1]
							inPckLen += 1

							buffers[i] = append(buffer[:bufferLen-1], buffer[bufferLen:]...)
						}
					}

					outPck, errPck := n.action(proc, inPcks)
					if outPck == nil && errPck == nil {
						if inPckLen == len(inStreams) {
							for _, inPck := range inPcks {
								proc.Stack().Clear(inPck.ID())
							}
						} else {
							for i, inPck := range inPcks {
								if inPck != nil {
									buffers[i] = append(buffers[i][:bufferLen], buffers[i][bufferLen-1:]...)
									buffers[i][bufferLen-1] = inPck
								}
							}
							return
						}
					}

					inStreams = lo.Filter[*port.Stream](inStreams, func(_ *port.Stream, index int) bool {
						return inPcks[i] != nil
					})
					inPcks = lo.Filter[*packet.Packet](inPcks, func(item *packet.Packet, _ int) bool {
						return item != nil
					})

					sendPacket := func(outPck *packet.Packet, outStream *port.Stream) {
						for _, inPck := range inPcks {
							if outPck == inPck {
								outPck = packet.New(outPck.Payload())
							}
						}
						for _, inPck := range inPcks {
							proc.Stack().Link(inPck.ID(), outPck.ID())
						}
						for _, inStream := range inStreams {
							if outStream.Links() > 0 {
								proc.Stack().Push(outPck.ID(), inStream.ID())
							}
						}
						for _, inStream := range inStreams {
							if outStream.Links() > 0 {
								outStream.Send(outPck)
							} else {
								inStream.Send(outPck)
							}
						}
					}

					if errPck != nil {
						sendPacket(errPck, errStream)
					} else if outPck != nil {
						sendPacket(outPck, outStream)
					}
				}()
			}
		}(i, inStream)
	}
	wg.Wait()
}

func (n *ManyToOneNode) backward(proc *process.Process, outStream *port.Stream) {
	n.mu.RLock()
	defer n.mu.RUnlock()

	var inStreams []*port.Stream

	for {
		backPck, ok := <-outStream.Receive()
		if !ok {
			return
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
