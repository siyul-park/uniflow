package control

import (
	"sync"

	"github.com/siyul-park/uniflow/pkg/node"
	"github.com/siyul-park/uniflow/pkg/packet"
	"github.com/siyul-park/uniflow/pkg/port"
	"github.com/siyul-park/uniflow/pkg/primitive"
	"github.com/siyul-park/uniflow/pkg/process"
)

type FlowNode struct {
	inPort  *port.Port
	outPort *port.Port
	mu      sync.RWMutex
}

var _ node.Node = (*FlowNode)(nil)

func NewFlowNode() *FlowNode {
	n := &FlowNode{
		inPort:  port.New(),
		outPort: port.New(),
	}
	n.inPort.AddInitHook(port.InitHookFunc(n.forward))
	return n
}

func (n *FlowNode) Port(name string) (*port.Port, bool) {
	n.mu.RLock()
	defer n.mu.RUnlock()

	switch name {
	case node.PortIn:
		return n.inPort, true
	case node.PortOut:
		return n.outPort, true
	default:
	}

	return nil, false
}

func (n *FlowNode) Close() error {
	n.mu.Lock()
	defer n.mu.Unlock()

	n.inPort.Close()
	n.outPort.Close()

	return nil
}

func (n *FlowNode) forward(proc *process.Process) {
	n.mu.RLock()
	defer n.mu.RUnlock()

	inStream := n.inPort.Open(proc)
	outStream := n.outPort.Open(proc)

	for {
		inPck, ok := <-inStream.Receive()
		if !ok {
			return
		}

		if outStream.Links() == 0 {
			proc.Stack().Clear(inPck.ID())
			continue
		}

		inPayload := inPck.Payload()

		var outPayloads []primitive.Value
		if inPayload, ok := inPayload.(*primitive.Slice); !ok {
			outPayloads = []primitive.Value{inPayload}
		} else {
			outPayloads = inPayload.Values()
		}

		for _, outPayload := range outPayloads {
			outPck := packet.New(outPayload)
			proc.Stack().Link(inPck.ID(), outPck.ID())
			proc.Stack().Push(outPck.ID(), inStream.ID())
			outStream.Send(outPck)
		}
	}
}
