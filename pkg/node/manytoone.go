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
	action   func(*process.Process, []*packet.Packet) (*packet.Packet, *packet.Packet)
	recorder *packetRecorder
	inPorts  []*port.InPort
	outPort  *port.OutPort
	errPort  *port.OutPort
	mu       sync.RWMutex
}

type packetRecorder struct {
	queues map[*process.Process]*packetQueue
	mu     sync.RWMutex
}

type packetQueue struct {
	data   [][]*packet.Packet
	resume bool
	mu     sync.RWMutex
}

var _ Node = (*ManyToOneNode)(nil)

// NewManyToOneNode creates a new ManyToOneNode instance with the given action function.
func NewManyToOneNode(action func(*process.Process, []*packet.Packet) (*packet.Packet, *packet.Packet)) *ManyToOneNode {
	n := &ManyToOneNode{
		action: action,
		recorder: &packetRecorder{
			queues: make(map[*process.Process]*packetQueue),
		},
		outPort: port.NewOut(),
		errPort: port.NewOut(),
	}

	if n.action != nil {
		n.outPort.AddHandler(port.HandlerFunc(n.backward))
		n.errPort.AddHandler(port.HandlerFunc(n.catch))
	}

	return n
}

// In returns the input port with the specified name.
func (n *ManyToOneNode) In(name string) *port.InPort {
	n.mu.RLock()
	defer n.mu.RUnlock()

	if i, ok := IndexOfMultiPort(PortIn, name); ok {
		for j := 0; j <= i; j++ {
			if len(n.inPorts) <= j {
				inPort := port.NewIn()
				n.inPorts = append(n.inPorts, inPort)

				if n.action != nil {
					inPort.AddHandler(port.HandlerFunc(func(proc *process.Process) {
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

	return nil
}

func (n *ManyToOneNode) forward(proc *process.Process, index int) {
	n.mu.RLock()
	defer n.mu.RUnlock()

	inReader := n.inPorts[index].Open(proc)
	outWriter := n.outPort.Open(proc)
	errWriter := n.errPort.Open(proc)

	for {
		inPck, ok := <-inReader.Read()
		if !ok {
			return
		}

		n.recorder.provide(proc, index, inPck)
		n.recorder.consume(proc, func(inPcks []*packet.Packet) bool {
			for len(inPcks) < len(n.inPorts) {
				inPcks = append(inPcks, nil)
			}

			outPck, errPck := n.action(proc, inPcks)

			inPcks = lo.Filter(inPcks, func(item *packet.Packet, _ int) bool {
				return item != nil
			})

			if errPck != nil {
				for _, inPck := range inPcks {
					proc.Stack().Add(inPck, errPck)
				}
				if errWriter.Links() > 0 {
					errWriter.Write(errPck)
				} else {
					n.receive(proc, errPck)
				}
			} else if outPck != nil {
				for _, inPck := range inPcks {
					proc.Stack().Add(inPck, outPck)
				}
				outWriter.Write(outPck)
			} else {
				if len(inPcks) < len(n.inPorts) {
					return false
				}
				for _, inPck := range inPcks {
					proc.Stack().Clear(inPck)
				}
			}
			return true
		})
	}
}

func (n *ManyToOneNode) backward(proc *process.Process) {
	n.mu.RLock()
	defer n.mu.RUnlock()

	outWriter := n.outPort.Open(proc)

	for {
		backPck, ok := <-outWriter.Receive()
		if !ok {
			return
		}

		n.receive(proc, backPck)
	}
}

func (n *ManyToOneNode) catch(proc *process.Process) {
	n.mu.RLock()
	defer n.mu.RUnlock()

	errWriter := n.errPort.Open(proc)

	for {
		backPck, ok := <-errWriter.Receive()
		if !ok {
			return
		}

		n.receive(proc, backPck)
	}
}

func (n *ManyToOneNode) receive(proc *process.Process, backPck *packet.Packet) {
	inReaders := make([]*port.Reader, len(n.inPorts))
	for i, inPort := range n.inPorts {
		inReaders[i] = inPort.Open(proc)
	}

	costs := make([]int, len(inReaders))
	for i, inReader := range inReaders {
		costs[i] = inReader.Cost(backPck)
	}

	min := lo.Min(costs)
	for i, cost := range costs {
		if cost == min {
			inReaders[i].Receive(backPck)
		}
	}
}

func (r *packetRecorder) provide(proc *process.Process, index int, pck *packet.Packet) {
	r.mu.Lock()
	defer r.mu.Unlock()

	queue, ok := r.queues[proc]
	if !ok {
		queue = &packetQueue{}
		r.queues[proc] = queue

		go func() {
			<-proc.Done()

			r.mu.Lock()
			defer r.mu.Unlock()

			delete(r.queues, proc)
		}()
	}

	queue.provide(index, pck)
}

func (r *packetRecorder) consume(proc *process.Process, fn func([]*packet.Packet) bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	if queue, ok := r.queues[proc]; ok {
		queue.consume(fn)
	}
}

func (q *packetQueue) provide(index int, pck *packet.Packet) {
	q.mu.Lock()
	defer q.mu.Unlock()

	for len(q.data) <= index {
		q.data = append(q.data, nil)
	}
	q.data[index] = append(q.data[index], pck)

	if !q.resume {
		q.resume = len(q.data[index]) == 1
	}
}

func (q *packetQueue) consume(fn func([]*packet.Packet) bool) {
	q.mu.Lock()
	defer q.mu.Unlock()

	if !q.resume {
		return
	}

	buffer := make([]*packet.Packet, len(q.data))
	for i, data := range q.data {
		if len(data) > 0 {
			buffer[i] = data[0]
		}
	}

	if fn(buffer) {
		for i := range q.data {
			if len(q.data[i]) > 0 {
				q.data[i] = q.data[i][1:]
			}
		}
	} else {
		q.resume = false
	}
}
