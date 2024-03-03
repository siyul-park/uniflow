package control

import (
	"sync"

	"github.com/siyul-park/uniflow/pkg/node"
	"github.com/siyul-park/uniflow/pkg/packet"
	"github.com/siyul-park/uniflow/pkg/port"
	"github.com/siyul-park/uniflow/pkg/primitive"
	"github.com/siyul-park/uniflow/pkg/process"
)

type IterateNode struct {
	inPort   *port.InPort
	outPorts []*port.OutPort
	mu       sync.RWMutex
}

var _ node.Node = (*IterateNode)(nil)

func NewIterateNode() *IterateNode {
	n := &IterateNode{
		inPort:   port.NewIn(),
		outPorts: []*port.OutPort{port.NewOut(), port.NewOut()},
	}

	n.inPort.AddHandler(port.HandlerFunc(n.forward))
	n.outPorts[0].AddHandler(port.HandlerFunc(n.backward))

	return n
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

	return nil
}

func (n *IterateNode) forward(proc *process.Process) {
	n.mu.RLock()
	defer n.mu.RUnlock()

	inReader := n.inPort.Open(proc)
	outWriter0 := n.outPorts[0].Open(proc)
	outWriter1 := n.outPorts[1].Open(proc)

	for {
		inPck, ok := <-inReader.Read()
		if !ok {
			return
		}

		inPayload := inPck.Payload()
		outPayloads := n.slice(inPayload)

		backPayloads := make([]primitive.Value, len(outPayloads))
		for i, outPayload := range outPayloads {
			outPck := packet.New(outPayload)
			proc.Stack().Add(inPck, outPck)

			outWriter0.Write(outPck)

			select {
			case <-proc.Stack().Done(outPck):
			case backPck, ok := <-outWriter0.Receive():
				if !ok {
					return
				}

				backPayloads[i] = backPck.Payload()
				proc.Stack().Unwind(backPck, outPck)
			}
		}

		outPayload := primitive.NewSlice(backPayloads...)
		outPck := packet.New(outPayload)
		proc.Stack().Add(inPck, outPck)

		if outWriter1.Links() > 0 {
			outWriter1.Write(outPck)
		} else {
			inReader.Receive(outPck)
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

func (n *IterateNode) slice(val primitive.Value) []primitive.Value {
	switch v := val.(type) {
	case *primitive.Slice:
		return v.Values()
	}
	return []primitive.Value{val}
}
