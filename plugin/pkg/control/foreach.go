package control

import (
	"slices"
	"sync"

	"github.com/gofrs/uuid"
	"github.com/siyul-park/uniflow/pkg/node"
	"github.com/siyul-park/uniflow/pkg/packet"
	"github.com/siyul-park/uniflow/pkg/port"
	"github.com/siyul-park/uniflow/pkg/primitive"
	"github.com/siyul-park/uniflow/pkg/process"
)

type ForeachNode struct {
	depth   int
	ioPort  *port.Port
	inPort  *port.Port
	outPort *port.Port
	errPort *port.Port
	mu      sync.RWMutex
}

var _ node.Node = (*ForeachNode)(nil)

func NewForeachNode(depth int) *ForeachNode {
	n := &ForeachNode{
		depth:   depth,
		ioPort:  port.New(),
		inPort:  port.New(),
		outPort: port.New(),
		errPort: port.New(),
	}

	n.inPort.AddInitHook(port.InitHookFunc(n.forward))
	n.outPort.AddInitHook(port.InitHookFunc(n.backward))

	return n
}

func (n *ForeachNode) Port(name string) *port.Port {
	n.mu.RLock()
	defer n.mu.RUnlock()

	switch name {
	case node.PortIO:
		return n.ioPort
	case node.PortIn:
		return n.inPort
	case node.PortOut:
		return n.outPort
	default:
	}

	return nil
}

// Close closes all.
func (n *ForeachNode) Close() error {
	n.mu.Lock()
	defer n.mu.Unlock()

	n.ioPort.Close()
	n.inPort.Close()
	n.outPort.Close()

	return nil
}

func (n *ForeachNode) forward(proc *process.Process) {
	n.mu.RLock()
	defer n.mu.RUnlock()

	ioStream := n.ioPort.Open(proc)
	inStream := n.inPort.Open(proc)
	outStream := n.outPort.Open(proc)

	ioStream.AddSendHook(port.SendHookFunc(func(pck *packet.Packet) {
		proc.Stack().Push(pck.ID(), ioStream.ID())
	}))
	outStream.AddSendHook(port.SendHookFunc(func(pck *packet.Packet) {
		proc.Stack().Push(pck.ID(), outStream.ID())
	}))

	backPcks := make(map[*packet.Packet][]*packet.Packet)
	go func() {
		for {
			backPck, ok := <-ioStream.Receive()
			if !ok {
				return
			}

			if _, ok := proc.Stack().Pop(backPck.ID(), ioStream.ID()); !ok {
				continue
			}

			for inPck := range backPcks {
				if slices.Contains[[]uuid.UUID, uuid.UUID](proc.Graph().Stems(backPck.ID()), inPck.ID()) {
					backPcks[inPck] = append(backPcks[inPck], backPck)
					break
				}
			}
		}
	}()

	for {
		inPck, ok := <-inStream.Receive()
		if !ok {
			return
		}

		inPayload := inPck.Payload()
		inPayloads := n.slice(inPayload, n.depth)

		var outPcks []*packet.Packet
		for _, inPayload := range inPayloads {
			outPck := packet.New(inPayload)
			proc.Graph().Add(inPck.ID(), outPck.ID())
			ioStream.Send(outPck)

			outPcks = append(outPcks, outPck)
		}

		backPcks[inPck] = nil

		go func() {
			for _, outPck := range outPcks {
				<-proc.Stack().Done(outPck.ID())
			}

			outPcks := backPcks[inPck]
			delete(backPcks, inPck)

			var outPayloads []primitive.Value
			for _, outPck := range outPcks {
				outPayloads = append(outPayloads, outPck.Payload())
			}

			outPck := packet.New(primitive.NewSlice(outPayloads...))
			proc.Graph().Add(inPck.ID(), outPck.ID())
			outStream.Send(outPck)
		}()
	}
}

func (n *ForeachNode) backward(proc *process.Process) {
	n.mu.RLock()
	defer n.mu.RUnlock()

	outStream := n.outPort.Open(proc)
	var inStream *port.Stream

	for {
		backPck, ok := <-outStream.Receive()
		if !ok {
			return
		}

		if _, ok := proc.Stack().Pop(backPck.ID(), outStream.ID()); !ok {
			continue
		}

		if inStream == nil {
			inStream = n.inPort.Open(proc)
		}

		inStream.Send(backPck)
	}
}

func (n *ForeachNode) slice(val primitive.Value, depth int) []primitive.Value {
	if depth == 0 {
		return []primitive.Value{val}
	}

	switch val := val.(type) {
	case *primitive.Slice:
		var elements []primitive.Value
		for i := 0; i < val.Len(); i++ {
			elements = append(elements, n.slice(val.Get(i), depth-1)...)
		}
		return elements
	}
	return []primitive.Value{val}
}
