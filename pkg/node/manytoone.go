package node

import (
	"sync"

	"github.com/siyul-park/uniflow/pkg/packet"
	"github.com/siyul-park/uniflow/pkg/port"
	"github.com/siyul-park/uniflow/pkg/process"
)

// ManyToOneNode processes packets from multiple input ports and sends them to one output port.
type ManyToOneNode struct {
	action     func(*process.Process, []*packet.Packet) (*packet.Packet, *packet.Packet)
	readGroups *process.Local[*packet.ReadGroup]
	tracer     *packet.Tracer
	inPorts    []*port.InPort
	outPort    *port.OutPort
	errPort    *port.OutPort
	mu         sync.RWMutex
}

var _ Node = (*ManyToOneNode)(nil)

// NewManyToOneNode creates a ManyToOneNode with the specified action function.
func NewManyToOneNode(action func(*process.Process, []*packet.Packet) (*packet.Packet, *packet.Packet)) *ManyToOneNode {
	n := &ManyToOneNode{
		action:     action,
		readGroups: process.NewLocal[*packet.ReadGroup](),
		tracer:     packet.NewTracer(),
		outPort:    port.NewOut(),
		errPort:    port.NewOut(),
	}

	if n.action != nil {
		n.outPort.AddListener(port.ListenFunc(n.backward))
		n.errPort.AddListener(port.ListenFunc(n.catch))
	}

	return n
}

// Out returns the output port with the specified name.
func (n *ManyToOneNode) In(name string) *port.InPort {
	n.mu.Lock()
	defer n.mu.Unlock()

	if NameOfPort(name) == PortIn {
		index, _ := IndexOfPort(name)
		for i := 0; i <= index; i++ {
			if len(n.inPorts) <= i {
				inPort := port.NewIn()
				n.inPorts = append(n.inPorts, inPort)
				if n.action != nil {
					i := i
					inPort.AddListener(port.ListenFunc(func(proc *process.Process) {
						n.forward(proc, i)
					}))
				}
			}
		}
		return n.inPorts[index]
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

// Close closes all ports and releases resources.
func (n *ManyToOneNode) Close() error {
	n.mu.RLock()
	defer n.mu.RUnlock()

	for _, inPort := range n.inPorts {
		inPort.Close()
	}
	n.outPort.Close()
	n.errPort.Close()
	n.readGroups.Close()
	n.tracer.Close()

	return nil
}

func (n *ManyToOneNode) forward(proc *process.Process, index int) {
	n.mu.RLock()
	defer n.mu.RUnlock()

	inReaders := make([]*packet.Reader, len(n.inPorts))
	for i, inPort := range n.inPorts {
		inReaders[i] = inPort.Open(proc)
	}
	outWriter := n.outPort.Open(proc)
	errWriter := n.errPort.Open(proc)

	readGroup, _ := n.readGroups.LoadOrStore(proc, func() (*packet.ReadGroup, error) {
		return packet.NewReadGroup(inReaders), nil
	})

	for inPck := range inReaders[index].Read() {
		n.tracer.Read(inReaders[index], inPck)

		if inPcks := readGroup.Read(inReaders[index], inPck); len(inPcks) < len(inReaders) {
			n.tracer.Transform(inPck, packet.None)
		} else if outPck, errPck := n.action(proc, inPcks); errPck != nil {
			n.tracer.Transform(inPck, errPck)
			n.tracer.Write(errWriter, errPck)
		} else if outPck != nil {
			n.tracer.Transform(inPck, outPck)
			n.tracer.Write(outWriter, outPck)
		} else {
			n.tracer.Transform(inPck, packet.None)
		}
	}
}

func (n *ManyToOneNode) backward(proc *process.Process) {
	n.mu.RLock()
	defer n.mu.RUnlock()

	outWriter := n.outPort.Open(proc)

	for backPck := range outWriter.Receive() {
		n.tracer.Receive(outWriter, backPck)
	}
}

func (n *ManyToOneNode) catch(proc *process.Process) {
	n.mu.RLock()
	defer n.mu.RUnlock()

	errWriter := n.errPort.Open(proc)

	for backPck := range errWriter.Receive() {
		n.tracer.Receive(errWriter, backPck)
	}
}
