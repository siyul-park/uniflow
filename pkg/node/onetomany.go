package node

import (
	"sync"

	"github.com/samber/lo"
	"github.com/siyul-park/uniflow/pkg/packet"
	"github.com/siyul-park/uniflow/pkg/port"
	"github.com/siyul-park/uniflow/pkg/process"
)

// OneToManyNode represents a node with one input and multiple output ports.
type OneToManyNode struct {
	action   func(*process.Process, *packet.Packet) ([]*packet.Packet, *packet.Packet)
	inPort   *port.InPort
	outPorts []*port.OutPort
	errPort  *port.OutPort
	mu       sync.RWMutex
}

var _ Node = (*OneToManyNode)(nil)

// NewOneToManyNode creates a new OneToManyNode instance with the given action function.
func NewOneToManyNode(action func(*process.Process, *packet.Packet) ([]*packet.Packet, *packet.Packet)) *OneToManyNode {
	n := &OneToManyNode{
		action:   action,
		inPort:   port.NewIn(),
		outPorts: nil,
		errPort:  port.NewOut(),
	}

	if n.action != nil {
		n.inPort.AddHandler(port.HandlerFunc(n.forward))
		n.errPort.AddHandler(port.HandlerFunc(n.catch))
	}

	return n
}

// In returns the input port.
func (n *OneToManyNode) In(name string) *port.InPort {
	n.mu.RLock()
	defer n.mu.RUnlock()

	switch name {
	case PortIn:
		return n.inPort
	}

	return nil
}

// Out returns the output port with the specified name.
func (n *OneToManyNode) Out(name string) *port.OutPort {
	n.mu.Lock()
	defer n.mu.Unlock()

	switch name {
	case PortErr:
		return n.errPort
	default:
		if i, ok := IndexOfPort(PortOut, name); ok {
			for j := 0; j <= i; j++ {
				if len(n.outPorts) <= j {
					outPort := port.NewOut()
					n.outPorts = append(n.outPorts, outPort)

					if n.action != nil {
						outPort.AddHandler(port.HandlerFunc(func(proc *process.Process) {
							n.backward(proc, j)
						}))
					}
				}
			}

			return n.outPorts[i]
		}
	}

	return nil
}

// Close closes all ports associated with the node.
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

	inReader := n.inPort.Open(proc)
	outWriters := make([]*port.Writer, len(n.outPorts))
	for i, outPort := range n.outPorts {
		outWriters[i] = outPort.Open(proc)
	}
	errWriter := n.errPort.Open(proc)

	for {
		inPck, ok := <-inReader.Read()
		if !ok {
			return
		}

		if outPcks, errPck := n.action(proc, inPck); errPck != nil {
			proc.Stack().Add(inPck, errPck)
			if errWriter.Links() > 0 {
				errWriter.Write(errPck)
			} else {
				inReader.Receive(errPck)
			}
		} else {
			if len(outPcks) > len(outWriters) {
				outPcks = outPcks[:len(outWriters)]
			}
			outWriters = lo.Filter(outWriters, func(_ *port.Writer, i int) bool {
				return len(outPcks) > i && outPcks[i] != nil
			})
			outPcks = lo.Filter(outPcks, func(item *packet.Packet, _ int) bool {
				return item != nil
			})

			if len(outPcks) > 0 {
				for _, outPck := range outPcks {
					proc.Stack().Add(inPck, outPck)
				}
				for i, outPck := range outPcks {
					outWriters[i].Write(outPck)
				}
			} else {
				proc.Stack().Clear(inPck)
			}
		}
	}
}

func (n *OneToManyNode) backward(proc *process.Process, index int) {
	n.mu.RLock()
	defer n.mu.RUnlock()

	inReader := n.inPort.Open(proc)
	outWriter := n.outPorts[index].Open(proc)

	for {
		backPck, ok := <-outWriter.Receive()
		if !ok {
			return
		}

		inReader.Receive(backPck)
	}
}

func (n *OneToManyNode) catch(proc *process.Process) {
	n.mu.RLock()
	defer n.mu.RUnlock()

	inReader := n.inPort.Open(proc)
	errWriter := n.errPort.Open(proc)

	for {
		backPck, ok := <-errWriter.Receive()
		if !ok {
			return
		}

		inReader.Receive(backPck)
	}
}
