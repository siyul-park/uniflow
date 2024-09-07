package node

import (
	"sync"

	"github.com/siyul-park/uniflow/pkg/packet"
	"github.com/siyul-park/uniflow/pkg/port"
	"github.com/siyul-park/uniflow/pkg/process"
)

// OneToManyNode processes a packet from one input port and sends it to multiple output ports.
type OneToManyNode struct {
	action   func(*process.Process, *packet.Packet) ([]*packet.Packet, *packet.Packet)
	tracer   *packet.Tracer
	inPort   *port.InPort
	outPorts []*port.OutPort
	errPort  *port.OutPort
	mu       sync.RWMutex
}

var _ Node = (*OneToManyNode)(nil)

// NewOneToManyNode creates a OneToManyNode with the specified action function.
func NewOneToManyNode(action func(*process.Process, *packet.Packet) ([]*packet.Packet, *packet.Packet)) *OneToManyNode {
	n := &OneToManyNode{
		action:   action,
		tracer:   packet.NewTracer(),
		inPort:   port.NewIn(),
		outPorts: nil,
		errPort:  port.NewOut(),
	}

	if n.action != nil {
		n.inPort.AddListener(port.ListenFunc(n.forward))
		n.errPort.AddListener(port.ListenFunc(n.catch))
	}

	return n
}

// In returns the input port with the specified name.
func (n *OneToManyNode) In(name string) *port.InPort {
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
func (n *OneToManyNode) Out(name string) *port.OutPort {
	n.mu.Lock()
	defer n.mu.Unlock()

	if name == PortErr {
		return n.errPort
	}

	if NameOfPort(name) == PortOut {
		index, _ := IndexOfPort(name)
		for i := 0; i <= index; i++ {
			if len(n.outPorts) <= i {
				outPort := port.NewOut()
				n.outPorts = append(n.outPorts, outPort)
				if n.action != nil {
					i := i
					outPort.AddListener(port.ListenFunc(func(proc *process.Process) {
						n.backward(proc, i)
					}))
				}
			}
		}
		return n.outPorts[index]
	}

	return nil
}

// Close closes all ports and releases resources.
func (n *OneToManyNode) Close() error {
	n.mu.RLock()
	defer n.mu.RUnlock()

	n.inPort.Close()
	for _, outPort := range n.outPorts {
		outPort.Close()
	}
	n.errPort.Close()
	n.tracer.Close()

	return nil
}

func (n *OneToManyNode) forward(proc *process.Process) {
	n.mu.RLock()
	defer n.mu.RUnlock()

	inReader := n.inPort.Open(proc)
	outWriters := make([]*packet.Writer, len(n.outPorts))
	for i, outPort := range n.outPorts {
		outWriters[i] = outPort.Open(proc)
	}
	errWriter := n.errPort.Open(proc)

	for inPck := range inReader.Read() {
		n.tracer.Read(inReader, inPck)

		if outPcks, errPck := n.action(proc, inPck); errPck != nil {
			n.tracer.Transform(inPck, errPck)
			n.tracer.Write(errWriter, errPck)
		} else {
			count := 0
			for i, outPck := range outPcks {
				if i < len(outWriters) && outPck != nil {
					n.tracer.Transform(inPck, outPck)
					count++
				}
			}
			if count > 0 {
				for i, outPck := range outPcks {
					if i < len(outWriters) && outPck != nil {
						n.tracer.Write(outWriters[i], outPck)
					}
				}
			} else {
				n.tracer.Transform(inPck, packet.None)
			}
		}
	}
}

func (n *OneToManyNode) backward(proc *process.Process, index int) {
	n.mu.RLock()
	defer n.mu.RUnlock()

	outWriter := n.outPorts[index].Open(proc)

	for backPck := range outWriter.Receive() {
		n.tracer.Receive(outWriter, backPck)
	}
}

func (n *OneToManyNode) catch(proc *process.Process) {
	n.mu.RLock()
	defer n.mu.RUnlock()

	errWriter := n.errPort.Open(proc)

	for backPck := range errWriter.Receive() {
		n.tracer.Receive(errWriter, backPck)
	}
}
