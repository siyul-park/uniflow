package node

import (
	"sync"

	"github.com/siyul-park/uniflow/pkg/packet"
	"github.com/siyul-park/uniflow/pkg/port"
	"github.com/siyul-park/uniflow/pkg/process"
)

// OneToOneNode represents a node with one input and one output port.
type OneToOneNode struct {
	action  func(*process.Process, *packet.Packet) (*packet.Packet, *packet.Packet)
	bridges *process.Local[*port.Bridge]
	inPort  *port.InPort
	outPort *port.OutPort
	errPort *port.OutPort
	mu      sync.RWMutex
}

var _ Node = (*OneToOneNode)(nil)

// NewOneToOneNode creates a new OneToOneNode instance with the given action function.
func NewOneToOneNode(action func(*process.Process, *packet.Packet) (*packet.Packet, *packet.Packet)) *OneToOneNode {
	n := &OneToOneNode{
		action:  action,
		bridges: process.NewLocal[*port.Bridge](),
		inPort:  port.NewIn(),
		outPort: port.NewOut(),
		errPort: port.NewOut(),
	}

	if n.action != nil {
		n.inPort.AddInitHook(port.InitHookFunc(n.forward))
		n.outPort.AddInitHook(port.InitHookFunc(n.backward))
		n.errPort.AddInitHook(port.InitHookFunc(n.catch))
	}

	return n
}

// In returns the input port with the specified name.
func (n *OneToOneNode) In(name string) *port.InPort {
	n.mu.RLock()
	defer n.mu.RUnlock()

	switch name {
	case PortIn:
		return n.inPort
	default:
	}

	return nil
}

// Out returns the output port with the specified name.
func (n *OneToOneNode) Out(name string) *port.OutPort {
	n.mu.RLock()
	defer n.mu.RUnlock()

	switch name {
	case PortOut:
		return n.outPort
	case PortErr:
		return n.errPort
	default:
	}

	return nil
}

// Close closes all ports associated with the node.
func (n *OneToOneNode) Close() error {
	n.mu.Lock()
	defer n.mu.Unlock()

	n.inPort.Close()
	n.outPort.Close()
	n.errPort.Close()
	n.bridges.Close()

	return nil
}

func (n *OneToOneNode) forward(proc *process.Process) {
	n.mu.RLock()
	defer n.mu.RUnlock()

	bridge, _ := n.bridges.LoadOrStore(proc, func() (*port.Bridge, error) {
		return port.NewBridge(), nil
	})

	inReader := n.inPort.Open(proc)
	outWriter := n.outPort.Open(proc)
	errWriter := n.errPort.Open(proc)

	for {
		inPck, ok := <-inReader.Read()
		if !ok {
			return
		}

		if outPck, errPck := n.action(proc, inPck); errPck != nil {
			bridge.Write([]*packet.Packet{errPck}, []*port.Reader{inReader}, []*port.Writer{errWriter})
		} else {
			bridge.Write([]*packet.Packet{outPck}, []*port.Reader{inReader}, []*port.Writer{outWriter})
		}
	}
}

func (n *OneToOneNode) backward(proc *process.Process) {
	n.mu.RLock()
	defer n.mu.RUnlock()

	bridge, _ := n.bridges.LoadOrStore(proc, func() (*port.Bridge, error) {
		return port.NewBridge(), nil
	})

	outWriter := n.outPort.Open(proc)

	for {
		backPck, ok := <-outWriter.Receive()
		if !ok {
			return
		}

		bridge.Receive(backPck, outWriter)
	}
}

func (n *OneToOneNode) catch(proc *process.Process) {
	n.mu.RLock()
	defer n.mu.RUnlock()

	bridge, _ := n.bridges.LoadOrStore(proc, func() (*port.Bridge, error) {
		return port.NewBridge(), nil
	})

	errWriter := n.errPort.Open(proc)

	for {
		backPck, ok := <-errWriter.Receive()
		if !ok {
			return
		}

		bridge.Receive(backPck, errWriter)
	}
}
