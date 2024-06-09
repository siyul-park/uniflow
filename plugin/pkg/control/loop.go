package control

import (
	"sync"

	"github.com/siyul-park/uniflow/pkg/node"
	"github.com/siyul-park/uniflow/pkg/object"
	"github.com/siyul-park/uniflow/pkg/packet"
	"github.com/siyul-park/uniflow/pkg/port"
	"github.com/siyul-park/uniflow/pkg/process"
	"github.com/siyul-park/uniflow/pkg/spec"
)

// LoopNode represents a node that Loops over input data in batches.
type LoopNode struct {
	bridges  *process.Local[*packet.Bridge]
	inPort   *port.InPort
	outPorts []*port.OutPort
	errPort  *port.OutPort
	mu       sync.RWMutex
}

// LoopNodeSpec holds the specifications for creating a LoopNode.
type LoopNodeSpec struct {
	spec.Meta `map:",inline"`
}

const KindLoop = "loop"

// NewLoopNode creates a new LoopNode with default configurations.
func NewLoopNode() *LoopNode {
	n := &LoopNode{
		bridges:  process.NewLocal[*packet.Bridge](),
		inPort:   port.NewIn(),
		outPorts: []*port.OutPort{port.NewOut(), port.NewOut()},
		errPort:  port.NewOut(),
	}

	n.inPort.AddInitHook(port.InitHookFunc(n.forward))
	n.outPorts[1].AddInitHook(port.InitHookFunc(n.backward))

	return n
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

	bridge, _ := n.bridges.LoadOrStore(proc, func() (*packet.Bridge, error) {
		return packet.NewBridge(), nil
	})

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

		var outPayloads []object.Object
		if v, ok := inPayload.(*object.Slice); ok {
			outPayloads = v.Values()
		} else {
			outPayloads = []object.Object{inPayload}
		}

		count := 0
		for _, outPayload := range outPayloads {
			outPck := packet.New(outPayload)
			if outWriter0.Write(outPck) > 0 {
				count++
			}
		}

		backPcks := make([]*packet.Packet, count)
		for i := 0; i < count; i++ {
			backPck, ok := <-outWriter0.Receive()
			if !ok {
				return
			}

			if _, ok := backPck.Payload().(*object.Error); ok {
				backPck = packet.CallOrFallback(errWriter, backPck, backPck)
			}
			backPcks[i] = backPck
		}

		backPck := packet.Merge(backPcks)
		if _, ok := backPck.Payload().(*object.Error); ok {
			bridge.Write(nil, []*packet.Reader{inReader}, nil)
		} else {
			bridge.Write([]*packet.Packet{backPck}, []*packet.Reader{inReader}, []*packet.Writer{outWriter1})
		}
	}
}

func (n *LoopNode) backward(proc *process.Process) {
	n.mu.RLock()
	defer n.mu.RUnlock()

	bridge, _ := n.bridges.LoadOrStore(proc, func() (*packet.Bridge, error) {
		return packet.NewBridge(), nil
	})

	outWriter1 := n.outPorts[1].Open(proc)

	for {
		backPck, ok := <-outWriter1.Receive()
		if !ok {
			return
		}

		bridge.Receive(backPck, outWriter1)
	}
}

// NewLoopNodeCodec creates a new codec for LoopNodeSpec.
func NewLoopNodeCodec() spec.Codec {
	return spec.CodecWithType(func(spec *LoopNodeSpec) (node.Node, error) {
		return NewLoopNode(), nil
	})
}
