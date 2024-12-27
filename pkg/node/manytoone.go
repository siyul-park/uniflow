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

// In Out returns the input port with the specified name.
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
					inPort.AddListener(n.forward(i))
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
	case PortError:
		return n.errPort
	default:
		return nil
	}
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

func (n *ManyToOneNode) forward(index int) port.Listener {
	return port.ListenFunc(func(proc *process.Process) {
		n.mu.RLock()
		defer n.mu.RUnlock()

		inReader := n.inPorts[index].Open(proc)
		var outWriter *packet.Writer
		var errWriter *packet.Writer

		readGroup, _ := n.readGroups.LoadOrStore(proc, func() (*packet.ReadGroup, error) {
			inReaders := make([]*packet.Reader, len(n.inPorts))
			for i, inPort := range n.inPorts {
				inReaders[i] = inPort.Open(proc)
			}
			return packet.NewReadGroup(inReaders), nil
		})

		for inPck := range inReader.Read() {
			n.tracer.Read(inReader, inPck)

			if outWriter == nil {
				outWriter = n.outPort.Open(proc)
			}
			if errWriter == nil {
				errWriter = n.errPort.Open(proc)
			}

			if inPcks := readGroup.Read(inReader, inPck); len(inPcks) < len(n.inPorts) {
				n.tracer.Reduce(inPck)
			} else if outPck, errPck := n.action(proc, inPcks); errPck != nil {
				n.tracer.Transform(inPck, errPck)
				n.tracer.Write(errWriter, errPck)
			} else if outPck != nil {
				n.tracer.Transform(inPck, outPck)
				n.tracer.Write(outWriter, outPck)
			} else {
				n.tracer.Reduce(inPck)
			}
		}
	})
}

func (n *ManyToOneNode) backward(proc *process.Process) {
	outWriter := n.outPort.Open(proc)

	for backPck := range outWriter.Receive() {
		n.tracer.Receive(outWriter, backPck)
	}
}

func (n *ManyToOneNode) catch(proc *process.Process) {
	errWriter := n.errPort.Open(proc)

	for backPck := range errWriter.Receive() {
		n.tracer.Receive(errWriter, backPck)
	}
}
