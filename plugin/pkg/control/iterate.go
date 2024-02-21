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
	inPort   *port.Port
	outPorts []*port.Port
	mu       sync.RWMutex
}

var _ node.Node = (*IterateNode)(nil)

func NewIterateNode(batch int) *IterateNode {
	if batch <= 0 {
		batch = 1
	}

	n := &IterateNode{
		batch:    batch,
		inPort:   port.New(),
		outPorts: []*port.Port{port.New(), port.New()},
	}

	n.inPort.AddInitHook(port.InitHookFunc(n.forward))
	n.outPorts[1].AddInitHook(port.InitHookFunc(n.backward))

	return n
}

func (n *IterateNode) Port(name string) *port.Port {
	n.mu.RLock()
	defer n.mu.RUnlock()

	switch name {
	case node.PortIn:
		return n.inPort
	default:
		if i, ok := node.IndexOfMultiPort(node.PortOut, name); ok {
			if i < len(n.outPorts) {
				return n.outPorts[i]
			}
		}
	}

	return nil
}

// Close closes all.
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

	inStream := n.inPort.Open(proc)
	loopStream := n.outPorts[0].Open(proc)
	doneStream := n.outPorts[1].Open(proc)

	loopStream.AddSendHook(port.SendHookFunc(func(pck *packet.Packet) {
		proc.Stack().Push(pck.ID(), loopStream.ID())
	}))
	doneStream.AddSendHook(port.SendHookFunc(func(pck *packet.Packet) {
		proc.Stack().Push(pck.ID(), doneStream.ID())
	}))

	for {
		inPck, ok := <-inStream.Receive()
		if !ok {
			return
		}

		inPayload := inPck.Payload()

		var outPayloads []primitive.Value
		if inPayloads, ok := inPayload.(*primitive.Slice); ok {
			if n.batch == 1 {
				outPayloads = inPayloads.Values()
			} else {
				for _, chunk := range lo.Chunk(inPayloads.Values(), n.batch) {
					outPayloads = append(outPayloads, primitive.NewSlice(chunk...))
				}
			}
		} else {
			outPayloads = append(outPayloads, inPayload)
		}

		var backPcks []*packet.Packet
		var errPck *packet.Packet
		for _, outPayload := range outPayloads {
			outPck := packet.New(outPayload)

			proc.Graph().Add(inPck.ID(), outPck.ID())
			proc.Stack().Push(outPck.ID(), inStream.ID())

			loopStream.Send(outPck)

			var backPck *packet.Packet
			for func() bool {
				select {
				case <-proc.Stack().Done(outPck.ID()):
					return false
				case backPck, ok = <-loopStream.Receive():
					if !ok {
						return false
					}
					if _, ok := proc.Stack().Pop(backPck.ID(), loopStream.ID()); !ok {
						return true
					}
					if head, ok := proc.Stack().Pop(backPck.ID(), inStream.ID()); !ok || head != outPck.ID() {
						return true
					}
					return false
				}
			}() {
				backPck = nil
			}

			if backPck != nil {
				if _, ok := packet.AsError(backPck); ok {
					errPck = backPck
					break
				}
				backPcks = append(backPcks, backPck)
			}
		}

		if errPck != nil {
			inStream.Send(errPck)
		} else if len(backPcks) > 0 {
			backPayloads := make([]primitive.Value, len(backPcks))
			for i, backPck := range backPcks {
				if backPck != nil {
					backPayloads[i] = backPck.Payload()
				}
			}

			outPayload := primitive.NewSlice(backPayloads...)
			outPck := packet.New(outPayload)
			for _, backPck := range backPcks {
				if backPck != nil {
					proc.Graph().Add(backPck.ID(), outPck.ID())
				}
			}

			if doneStream.Links() > 0 {
				doneStream.Send(outPck)
			} else {
				inStream.Send(outPck)
			}
		} else {
			proc.Stack().Clear(inPck.ID())
		}
	}
}

func (n *IterateNode) backward(proc *process.Process) {
	n.mu.RLock()
	defer n.mu.RUnlock()

	var inStream *port.Stream
	doneStream := n.outPorts[1].Open(proc)

	for {
		backPck, ok := <-doneStream.Receive()
		if !ok {
			return
		}

		if _, ok := proc.Stack().Pop(backPck.ID(), doneStream.ID()); !ok {
			continue
		}

		if inStream == nil {
			inStream = n.inPort.Open(proc)
		}

		inStream.Send(backPck)
	}
}
