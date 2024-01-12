package control

import (
	"sync"

	"github.com/oklog/ulid/v2"
	"github.com/siyul-park/uniflow/pkg/node"
	"github.com/siyul-park/uniflow/pkg/packet"
	"github.com/siyul-park/uniflow/pkg/port"
	"github.com/siyul-park/uniflow/pkg/primitive"
	"github.com/siyul-park/uniflow/pkg/process"
	"github.com/siyul-park/uniflow/pkg/scheme"
)

// FlowNode represents a node that processes packets in a flow-like manner.
type FlowNode struct {
	inPort  *port.Port
	outPort *port.Port
	mu      sync.RWMutex
}

// FlowNodeSpec holds the specifications for creating a FlowNode.
type FlowNodeSpec struct {
	scheme.SpecMeta
}

const KindFlow = "flow"

var _ node.Node = (*FlowNode)(nil)

// NewFlowNodeCodec creates a new codec for FlowNodeSpec.
func NewFlowNodeCodec() scheme.Codec {
	return scheme.CodecWithType[*FlowNodeSpec](func(_ *FlowNodeSpec) (node.Node, error) {
		return NewFlowNode(), nil
	})
}

// NewFlowNode creates a new FlowNode.
func NewFlowNode() *FlowNode {
	n := &FlowNode{
		inPort:  port.New(),
		outPort: port.New(),
	}

	n.inPort.AddInitHook(port.InitHookFunc(n.forward))
	n.outPort.AddInitHook(port.InitHookFunc(n.backward))

	return n
}

// Port returns the specified port.
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

// Close closes all.
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
		if inPayloads, ok := inPayload.(*primitive.Slice); !ok {
			outPayloads = []primitive.Value{inPayload}
		} else {
			outPayloads = inPayloads.Values()
		}

		var outPcks []*packet.Packet
		for _, outPayload := range outPayloads {
			outPck := packet.New(outPayload)
			proc.Stack().Link(inPck.ID(), outPck.ID())
			proc.Stack().Push(outPck.ID(), inStream.ID())
			outPcks = append(outPcks, outPck)
		}

		for _, outPck := range outPcks {
			outStream.Send(outPck)
		}
	}
}

func (n *FlowNode) backward(proc *process.Process) {
	n.mu.RLock()
	defer n.mu.RUnlock()

	var inStream *port.Stream
	outStream := n.outPort.Open(proc)

	buffers := make(map[ulid.ULID][]primitive.Value)

	for {
		backPck, ok := <-outStream.Receive()
		if !ok {
			return
		}

		if inStream == nil {
			inStream = n.inPort.Open(proc)
		}

		if heads, ok := proc.Stack().Pop(backPck.ID(), inStream.ID()); ok {
			for _, head := range heads {
				buffers[head] = append(buffers[head], backPck.Payload())
				if len(proc.Stack().Leaves(head)) == 0 {
					backPayload := primitive.NewSlice(buffers[head]...)
					backPck := packet.New(backPayload)

					proc.Stack().Link(head, backPck.ID())
					inStream.Send(backPck)

					delete(buffers, head)
				}
			}
		}
	}
}
