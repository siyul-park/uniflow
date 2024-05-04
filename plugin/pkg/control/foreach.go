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

// ForEachNode represents a node that ForEachs over input data in batches.
type ForEachNode struct {
	batch    int
	inPort   *port.InPort
	outPorts []*port.OutPort
	errPort  *port.OutPort
	mu       sync.RWMutex
}

// ForEachNodeSpec holds the specifications for creating a ForEachNode.
type ForEachNodeSpec struct {
	scheme.SpecMeta `map:",inline"`
	Batch           int `map:"batch,omitempty"`
}

const KindForEach = "foreach"

// NewForEachNode creates a new ForEachNode with default configurations.
func NewForEachNode() *ForEachNode {
	n := &ForEachNode{
		batch:    1,
		inPort:   port.NewIn(),
		outPorts: []*port.OutPort{port.NewOut(), port.NewOut()},
		errPort:  port.NewOut(),
	}

	n.inPort.AddHandler(port.HandlerFunc(n.forward))
	n.outPorts[0].AddHandler(port.HandlerFunc(n.backward))

	return n
}

// Batch returns the batch size of the ForEachNode.
func (n *ForEachNode) Batch() int {
	n.mu.RLock()
	defer n.mu.RUnlock()
	return n.batch
}

// SetBatch sets the batch size of the ForEachNode.
func (n *ForEachNode) SetBatch(batch int) {
	n.mu.Lock()
	defer n.mu.Unlock()

	if batch < 1 {
		batch = 1
	}
	n.batch = batch
}

// In returns the input port with the specified name.
func (n *ForEachNode) In(name string) *port.InPort {
	n.mu.RLock()
	defer n.mu.RUnlock()

	switch name {
	case node.PortIn:
		return n.inPort
	}

	return nil
}

// Out returns the output port with the specified name.
func (n *ForEachNode) Out(name string) *port.OutPort {
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
func (n *ForEachNode) Close() error {
	n.mu.Lock()
	defer n.mu.Unlock()

	n.inPort.Close()
	for _, p := range n.outPorts {
		p.Close()
	}
	n.errPort.Close()

	return nil
}

func (n *ForEachNode) forward(proc *process.Process) {
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
		outPayloads := n.slice(inPayload)

		outPcks := make([]*packet.Packet, len(outPayloads))
		for i, outPayload := range outPayloads {
			outPck := packet.New(outPayload)
			proc.Stack().Add(inPck, outPck)

			outPcks[i] = outPck
		}

		var backPcks []*packet.Packet
		var catch bool
	Loop:
		for i, outPck := range outPcks {
			if !outWriter0.Write(outPck) {
				proc.Stack().Clear(outPck)
			}

			select {
			case <-proc.Stack().Done(outPck):
			case backPck, ok := <-outWriter0.Receive():
				if !ok {
					return
				}

				if _, ok := packet.AsError(backPck); ok && errWriter.Links() > 0 {
					if !errWriter.Write(backPck) {
						proc.Stack().Clear(backPck)
					}
					if backPck, ok = <-errWriter.Receive(); !ok {
						return
					}
				}

				proc.Stack().Unwind(backPck, outPck)

				if _, ok := packet.AsError(backPck); ok {
					inReader.Receive(backPck)

					for j := 0; j < i; j++ {
						proc.Stack().Clear(outPcks[j])
					}
					for _, backPck := range backPcks {
						proc.Stack().Clear(backPck)
					}
					backPcks = nil
					catch = true

					break Loop
				}

				backPcks = append(backPcks, backPck)
			}
		}

		if len(backPcks) > 0 {
			backPayloads := lo.Map(backPcks, func(backPck *packet.Packet, _ int) primitive.Value {
				return backPck.Payload()
			})

			outPayload := primitive.NewSlice(backPayloads...)
			outPck := packet.New(outPayload)
			proc.Stack().Add(inPck, outPck)

			if !outWriter1.Write(outPck) {
				if !inReader.Receive(outPck) {
					proc.Stack().Clear(outPck)
				}
			}
		} else if !catch {
			proc.Stack().Clear(inPck)
		}
	}
}

func (n *ForEachNode) backward(proc *process.Process) {
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

func (n *ForEachNode) slice(val primitive.Value) []primitive.Value {
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

// NewForEachNodeCodec creates a new codec for ForEachNodeSpec.
func NewForEachNodeCodec() scheme.Codec {
	return scheme.CodecWithType(func(spec *ForEachNodeSpec) (node.Node, error) {
		n := NewForEachNode()
		n.SetBatch(spec.Batch)

		return n, nil
	})
}
