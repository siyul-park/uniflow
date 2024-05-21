package node

import (
	"sync"

	"github.com/siyul-park/uniflow/pkg/packet"
	"github.com/siyul-park/uniflow/pkg/port"
	"github.com/siyul-park/uniflow/pkg/process"
)

// OneToManyNode represents a node with one input and multiple output ports.
type OneToManyNode struct {
	action   func(*process.Process, *packet.Packet) ([]*packet.Packet, *packet.Packet)
	bridges  *process.Local[*port.Bridge]
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
		bridges:  process.NewLocal[*port.Bridge](),
		inPort:   port.NewIn(),
		outPorts: nil,
		errPort:  port.NewOut(),
	}

	if n.action != nil {
		n.inPort.AddInitHook(port.InitHookFunc(n.forward))
		n.errPort.AddInitHook(port.InitHookFunc(n.catch))
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
						outPort.AddInitHook(port.InitHookFunc(func(proc *process.Process) {
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
	n.bridges.Close()

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

	bridge, _ := n.bridges.LoadOrStore(proc, func() (*port.Bridge, error) {
		return port.NewBridge(), nil
	})

	for {
		inPck, ok := <-inReader.Read()
		if !ok {
			return
		}

		if outPcks, errPck := n.action(proc, inPck); errPck != nil {
			if errWriter.Write(errPck) == 0 {
				inReader.Receive(errPck)
			}
		} else {
			bridge.Write(outPcks, []*port.Reader{inReader}, outWriters)
		}
	}
}

func (n *OneToManyNode) backward(proc *process.Process, index int) {
	n.mu.RLock()
	defer n.mu.RUnlock()

	outWriter := n.outPorts[index].Open(proc)
	bridge, _ := n.bridges.LoadOrStore(proc, func() (*port.Bridge, error) {
		return port.NewBridge(), nil
	})

	for {
		backPck, ok := <-outWriter.Receive()
		if !ok {
			return
		}

		bridge.Receive(backPck, outWriter)
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
