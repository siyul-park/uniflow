package control

import (
	"sync"

	"github.com/samber/lo"
	"github.com/siyul-park/uniflow/pkg/node"
	"github.com/siyul-park/uniflow/pkg/packet"
	"github.com/siyul-park/uniflow/pkg/port"
	"github.com/siyul-park/uniflow/pkg/primitive"
	"github.com/siyul-park/uniflow/pkg/process"
)

type IterateNode struct {
	batch    int
	inPort   *port.InPort
	outPorts []*port.OutPort
	errPort  *port.OutPort
	mu       sync.RWMutex
}

var _ node.Node = (*IterateNode)(nil)

func NewIterateNode() *IterateNode {
	n := &IterateNode{
		batch:    1,
		inPort:   port.NewIn(),
		outPorts: []*port.OutPort{port.NewOut(), port.NewOut()},
		errPort:  port.NewOut(),
	}

	n.inPort.AddHandler(port.HandlerFunc(n.forward))
	n.outPorts[0].AddHandler(port.HandlerFunc(n.backward))

	return n
}

func (n *IterateNode) Batch() int {
	n.mu.RLock()
	defer n.mu.RUnlock()

	return n.batch
}

func (n *IterateNode) SetBatch(batch int) {
	n.mu.Lock()
	defer n.mu.Unlock()

	n.batch = batch
}

func (n *IterateNode) In(name string) *port.InPort {
	n.mu.RLock()
	defer n.mu.RUnlock()

	switch name {
	case node.PortIn:
		return n.inPort
	}

	return nil
}

func (n *IterateNode) Out(name string) *port.OutPort {
	n.mu.RLock()
	defer n.mu.RUnlock()

	switch name {
	case node.PortOut:
		return n.outPorts[0]
	case node.PortErr:
		return n.errPort
	default:
		if i, ok := node.IndexOfPort(node.PortOut, name); ok {
			if len(n.outPorts) > i {
				return n.outPorts[i]
			}
		}
	}

	return nil
}

func (n *IterateNode) Close() error {
	n.mu.Lock()
	defer n.mu.Unlock()

	n.inPort.Close()
	for _, p := range n.outPorts {
		p.Close()
	}
	n.errPort.Close()

	return nil
}

func (n *IterateNode) forward(proc *process.Process) {
	n.mu.RLock()
	defer n.mu.RUnlock()

	inReader := n.inPort.Open(proc)

	for {
		inPck, ok := <-inReader.Read()
		if !ok {
			return
		}

		inPayload := inPck.Payload()
		outPayloads := n.slice(inPayload)

		var backPcks []*packet.Packet
		for _, outPayload := range outPayloads {
			outPck := packet.New(outPayload)
			proc.Stack().Add(inPck, outPck)

			backPck, ok := n.loop(proc, outPck)

			if !ok {
				for _, backPck := range backPcks {
					proc.Stack().Clear(backPck)
				}
				backPcks = nil
				break
			}

			if backPck != nil {
				backPcks = append(backPcks, backPck)
			}
		}

		backPayloads := lo.Map(backPcks, func(backPck *packet.Packet, _ int) primitive.Value {
			return backPck.Payload()
		})

		if len(backPayloads) > 0 {
			outPayload := primitive.NewSlice(backPayloads...)
			outPck := packet.New(outPayload)
			proc.Stack().Add(inPck, outPck)

			n.receive(proc, outPck)
		} else {
			proc.Stack().Clear(inPck)
		}
	}
}

func (n *IterateNode) backward(proc *process.Process) {
	n.mu.RLock()
	defer n.mu.RUnlock()

	inReader := n.inPort.Open(proc)
	outWriter1 := n.outPorts[1].Open(proc)

	for {
		backPck, ok := <-outWriter1.Receive()
		if !ok {
			return
		}

		inReader.Receive(backPck)
	}
}

func (n *IterateNode) loop(proc *process.Process, outPck *packet.Packet) (*packet.Packet, bool) {
	inReader := n.inPort.Open(proc)
	outWriter0 := n.outPorts[0].Open(proc)

	outWriter0.Write(outPck)

	select {
	case <-proc.Stack().Done(outPck):
		return nil, true
	case backPck, ok := <-outWriter0.Receive():
		if !ok {
			return nil, false
		}

		if _, ok := packet.AsError(backPck); ok {
			backPck = n.catch(proc, backPck)
		}

		proc.Stack().Unwind(backPck, outPck)

		if _, ok := packet.AsError(backPck); ok {
			inReader.Receive(backPck)
			return nil, false
		}
		return backPck, true
	}
}

func (n *IterateNode) catch(proc *process.Process, errPck *packet.Packet) *packet.Packet {
	errWriter := n.errPort.Open(proc)

	if errWriter.Links() == 0 {
		return errPck
	}

	errWriter.Write(errPck)

	backPck, ok := <-errWriter.Receive()
	if !ok {
		return errPck
	}
	return backPck
}

func (n *IterateNode) receive(proc *process.Process, backPck *packet.Packet) {
	inReader := n.inPort.Open(proc)
	outWriter1 := n.outPorts[1].Open(proc)

	if outWriter1.Links() > 0 {
		outWriter1.Write(backPck)
	} else {
		inReader.Receive(backPck)
	}
}

func (n *IterateNode) slice(val primitive.Value) []primitive.Value {
	var values []primitive.Value

	switch v := val.(type) {
	case *primitive.Slice:
		values = v.Values()
	default:
		values = []primitive.Value{val}
	}

	if n.batch > 1 {
		chunks := lo.Chunk(values, n.batch)

		values = nil
		for _, chunk := range chunks {
			values = append(values, primitive.NewSlice(chunk...))
		}
	}

	return values
}
