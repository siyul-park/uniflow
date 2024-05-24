package control

import (
	"sync"

	"github.com/samber/lo"
	"github.com/siyul-park/uniflow/pkg/node"
	"github.com/siyul-park/uniflow/pkg/packet"
	"github.com/siyul-park/uniflow/pkg/port"
	"github.com/siyul-park/uniflow/pkg/primitive"
	"github.com/siyul-park/uniflow/pkg/process"
	"github.com/siyul-park/uniflow/pkg/scheme"
)

// LoopNode represents a node that Loops over input data in batches.
type LoopNode struct {
	batch    int
	bridges  *process.Local[*port.Bridge]
	inPort   *port.InPort
	outPorts []*port.OutPort
	errPort  *port.OutPort
	mu       sync.RWMutex
}

// LoopNodeSpec holds the specifications for creating a LoopNode.
type LoopNodeSpec struct {
	scheme.SpecMeta `map:",inline"`
	Batch           int `map:"batch,omitempty"`
}

const KindLoop = "loop"

// NewLoopNode creates a new LoopNode with default configurations.
func NewLoopNode() *LoopNode {
	n := &LoopNode{
		batch:    1,
		bridges:  process.NewLocal[*port.Bridge](),
		inPort:   port.NewIn(),
		outPorts: []*port.OutPort{port.NewOut(), port.NewOut()},
		errPort:  port.NewOut(),
	}

	n.inPort.AddInitHook(port.InitHookFunc(n.forward))

	return n
}

// Batch returns the batch size of the LoopNode.
func (n *LoopNode) Batch() int {
	n.mu.RLock()
	defer n.mu.RUnlock()
	return n.batch
}

// SetBatch sets the batch size of the LoopNode.
func (n *LoopNode) SetBatch(batch int) {
	n.mu.Lock()
	defer n.mu.Unlock()

	if batch < 1 {
		batch = 1
	}
	n.batch = batch
}

// In returns the input port with the specified name.
func (n *LoopNode) In(name string) *port.InPort {
	n.mu.RLock()
	defer n.mu.RUnlock()

	switch name {
	case node.PortIn:
		return n.inPort
	}

	return nil
}

// Out returns the output port with the specified name.
func (n *LoopNode) Out(name string) *port.OutPort {
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

// Close closes all ports associated with the node.
func (n *LoopNode) Close() error {
	n.mu.Lock()
	defer n.mu.Unlock()

	n.inPort.Close()
	for _, p := range n.outPorts {
		p.Close()
	}
	n.errPort.Close()
	n.bridges.Close()

	return nil
}

func (n *LoopNode) forward(proc *process.Process) {
	n.mu.RLock()
	defer n.mu.RUnlock()

	inReader := n.inPort.Open(proc)
	outWriter0 := n.outPorts[0].Open(proc)
	outWriter1 := n.outPorts[1].Open(proc)
	errWriter := n.errPort.Open(proc)

	for {
		inPck, ok := <-inReader.Read()
		if !ok {
			return
		}

		inPayload := inPck.Payload()
		outPayloads := n.chunk(inPayload)

		outPcks := make([]*packet.Packet, len(outPayloads))
		for i, outPayload := range outPayloads {
			outPck := packet.New(outPayload)
			outPcks[i] = outPck
		}

		backPcks := make([]*packet.Packet, len(outPcks))
		for i, outPck := range outPcks {
			backPcks[i] = port.Call(outWriter0, outPck)
		}

		backPck := packet.Merge(backPcks)
		if _, ok := packet.AsError(backPck); ok {
			if errWriter.Write(backPck) > 0 {
				backPck = <-errWriter.Receive()
			}
		}

		if _, ok := packet.AsError(backPck); !ok {
			backPck = port.Call(outWriter1, backPck)
		}

		inReader.Receive(backPck)
	}
}

func (n *LoopNode) chunk(val primitive.Value) []primitive.Value {
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

// NewLoopNodeCodec creates a new codec for LoopNodeSpec.
func NewLoopNodeCodec() scheme.Codec {
	return scheme.CodecWithType(func(spec *LoopNodeSpec) (node.Node, error) {
		n := NewLoopNode()
		n.SetBatch(spec.Batch)

		return n, nil
	})
}
