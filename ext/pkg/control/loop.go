package control

import (
	"sync"

	"github.com/siyul-park/uniflow/pkg/node"
	"github.com/siyul-park/uniflow/pkg/object"
	"github.com/siyul-park/uniflow/pkg/packet"
	"github.com/siyul-park/uniflow/pkg/port"
	"github.com/siyul-park/uniflow/pkg/process"
	"github.com/siyul-park/uniflow/pkg/scheme"
	"github.com/siyul-park/uniflow/pkg/spec"
)

// LoopNode represents a node that Loops over input data in batches.
type LoopNode struct {
	tracer   *packet.Tracer
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
		tracer:   packet.NewTracer(),
		inPort:   port.NewIn(),
		outPorts: []*port.OutPort{port.NewOut(), port.NewOut()},
		errPort:  port.NewOut(),
	}

	n.inPort.AddInitHook(port.InitHookFunc(n.forward))
	n.outPorts[0].AddInitHook(port.InitHookFunc(n.backward0))
	n.outPorts[1].AddInitHook(port.InitHookFunc(n.backward1))
	n.errPort.AddInitHook(port.InitHookFunc(n.catch))

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
	n.tracer.Close()

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
		n.tracer.Read(inReader, inPck)

		inPayload := inPck.Payload()

		var outPayloads []object.Object
		if v, ok := inPayload.(object.Slice); ok {
			outPayloads = v.Values()
		} else {
			outPayloads = []object.Object{inPayload}
		}

		outPcks := make([]*packet.Packet, len(outPayloads))
		for i, outPayload := range outPayloads {
			outPck := packet.New(outPayload)
			outPcks[i] = outPck
			n.tracer.Transform(inPck, outPck)
		}

		n.tracer.Sniff(inPck, packet.HandlerFunc(func(backPck *packet.Packet) {
			n.tracer.Transform(inPck, backPck)
			if _, ok := backPck.Payload().(object.Error); ok {
				n.tracer.Write(errWriter, backPck)
			} else {
				n.tracer.Write(outWriter1, backPck)
			}
		}))

		for _, outPck := range outPcks {
			n.tracer.Write(outWriter0, outPck)
		}
	}
}

func (n *LoopNode) backward0(proc *process.Process) {
	n.mu.RLock()
	defer n.mu.RUnlock()

	outWriter0 := n.outPorts[0].Open(proc)

	for {
		backPck, ok := <-outWriter0.Receive()
		if !ok {
			return
		}

		n.tracer.Receive(outWriter0, backPck)
	}
}

func (n *LoopNode) backward1(proc *process.Process) {
	n.mu.RLock()
	defer n.mu.RUnlock()

	outWriter1 := n.outPorts[1].Open(proc)

	for {
		backPck, ok := <-outWriter1.Receive()
		if !ok {
			return
		}

		n.tracer.Receive(outWriter1, backPck)
	}
}

func (n *LoopNode) catch(proc *process.Process) {
	n.mu.RLock()
	defer n.mu.RUnlock()

	errWriter := n.errPort.Open(proc)

	for {
		backPck, ok := <-errWriter.Receive()
		if !ok {
			return
		}

		n.tracer.Receive(errWriter, backPck)
	}
}

// NewLoopNodeCodec creates a new codec for LoopNodeSpec.
func NewLoopNodeCodec() scheme.Codec {
	return scheme.CodecWithType(func(spec *LoopNodeSpec) (node.Node, error) {
		return NewLoopNode(), nil
	})
}
