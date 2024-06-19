package node

import (
	"sync"

	"github.com/samber/lo"
	"github.com/siyul-park/uniflow/pkg/packet"
	"github.com/siyul-park/uniflow/pkg/port"
	"github.com/siyul-park/uniflow/pkg/process"
)

// ManyToOneNode represents a node with multiple input ports and one output port.
type ManyToOneNode struct {
	action      func(*process.Process, []*packet.Packet) (*packet.Packet, *packet.Packet)
	dispatchers *process.Local[*packet.Dispatcher]
	tracers     *process.Local[*packet.Tracer]
	inPorts     []*port.InPort
	outPort     *port.OutPort
	errPort     *port.OutPort
	mu          sync.RWMutex
}

var _ Node = (*ManyToOneNode)(nil)

// NewManyToOneNode creates a new ManyToOneNode instance with the given action function.
func NewManyToOneNode(action func(*process.Process, []*packet.Packet) (*packet.Packet, *packet.Packet)) *ManyToOneNode {
	n := &ManyToOneNode{
		action:      action,
		dispatchers: process.NewLocal[*packet.Dispatcher](),
		tracers:     process.NewLocal[*packet.Tracer](),
		outPort:     port.NewOut(),
		errPort:     port.NewOut(),
	}

	if n.action != nil {
		n.outPort.AddInitHook(port.InitHookFunc(n.backward))
		n.errPort.AddInitHook(port.InitHookFunc(n.catch))
	}

	return n
}

// In returns the input port with the specified name.
func (n *ManyToOneNode) In(name string) *port.InPort {
	n.mu.RLock()
	defer n.mu.RUnlock()

	if i, ok := IndexOfPort(PortIn, name); ok {
		for j := 0; j <= i; j++ {
			if len(n.inPorts) <= j {
				inPort := port.NewIn()
				n.inPorts = append(n.inPorts, inPort)

				if n.action != nil {
					j := j
					inPort.AddInitHook(port.InitHookFunc(func(proc *process.Process) {
						n.forward(proc, j)
					}))
				}
			}
		}

		return n.inPorts[i]
	}

	return nil
}

// Out returns the output port with the specified name.
func (n *ManyToOneNode) Out(name string) *port.OutPort {
	n.mu.RLock()
	defer n.mu.RUnlock()

	switch name {
	case PortOut:
		return n.outPort
	case PortErr:
		return n.errPort
	}

	return nil
}

// Close closes all ports associated with the node.
func (n *ManyToOneNode) Close() error {
	n.mu.Lock()
	defer n.mu.Unlock()

	for _, p := range n.inPorts {
		p.Close()
	}
	n.outPort.Close()
	n.errPort.Close()
	n.dispatchers.Close()
	n.tracers.Close()

	return nil
}

func (n *ManyToOneNode) forward(proc *process.Process, index int) {
	n.mu.RLock()
	defer n.mu.RUnlock()

	dispatcher, _ := n.dispatchers.LoadOrStore(proc, func() (*packet.Dispatcher, error) {
		tracer, _ := n.tracers.LoadOrStore(proc, func() (*packet.Tracer, error) {
			return packet.NewTracer(), nil
		})

		inReaders := make([]*packet.Reader, len(n.inPorts))
		for i, inPort := range n.inPorts {
			inReaders[i] = inPort.Open(proc)
		}
		outWriter := n.outPort.Open(proc)
		errWriter := n.errPort.Open(proc)

		return packet.NewDispatcher(inReaders, packet.RouteHookFunc(func(inPcks []*packet.Packet) bool {
			inReaders := lo.Filter(inReaders, func(_ *packet.Reader, i int) bool {
				return inPcks[i] != nil
			})

			outPck, errPck := n.action(proc, inPcks)

			if outPck == nil && errPck == nil {
				if len(inPcks) == len(inReaders) {
					for i, inReader := range inReaders {
						tracer.Read(inReader, inPcks[i])
					}
					for _, inPck := range inPcks {
						tracer.Transform(inPck, packet.None)
					}
					return true
				}
				return false
			}

			for i, inReader := range inReaders {
				tracer.Read(inReader, inPcks[i])
			}

			if errPck != nil {
				for _, inPck := range inPcks {
					tracer.Transform(inPck, errPck)
				}
				tracer.Write(errWriter, errPck)
			} else {
				for _, inPck := range inPcks {
					tracer.Transform(inPck, outPck)
				}
				tracer.Write(outWriter, outPck)
			}
			return true
		})), nil
	})

	inReader := n.inPorts[index].Open(proc)

	for {
		inPck, ok := <-inReader.Read()
		if !ok {
			return
		}

		dispatcher.Write(inPck, inReader)
	}
}

func (n *ManyToOneNode) backward(proc *process.Process) {
	n.mu.RLock()
	defer n.mu.RUnlock()

	tracer, _ := n.tracers.LoadOrStore(proc, func() (*packet.Tracer, error) {
		return packet.NewTracer(), nil
	})

	outWriter := n.outPort.Open(proc)

	for {
		backPck, ok := <-outWriter.Receive()
		if !ok {
			return
		}

		tracer.Receive(outWriter, backPck)
	}
}

func (n *ManyToOneNode) catch(proc *process.Process) {
	n.mu.RLock()
	defer n.mu.RUnlock()

	tracer, _ := n.tracers.LoadOrStore(proc, func() (*packet.Tracer, error) {
		return packet.NewTracer(), nil
	})

	errWriter := n.errPort.Open(proc)

	for {
		backPck, ok := <-errWriter.Receive()
		if !ok {
			return
		}

		tracer.Receive(errWriter, backPck)
	}
}
