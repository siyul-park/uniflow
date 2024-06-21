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

// CallNode redirects packets from the input port to the intermediate port for processing by connected nodes, then outputs the results to the output port.
type CallNode struct {
	tracers  *process.Local[*packet.Tracer]
	inPort   *port.InPort
	outPorts []*port.OutPort
	errPort  *port.OutPort
	mu       sync.RWMutex
}

// CallNodeSpec holds the specifications for creating a CallNode.
type CallNodeSpec struct {
	spec.Meta `map:",inline"`
}

var _ node.Node = (*CallNode)(nil)

const KindCall = "call"

// NewCallNode creates a new CallNode.
func NewCallNode() *CallNode {
	n := &CallNode{
		tracers:  process.NewLocal[*packet.Tracer](),
		inPort:   port.NewIn(),
		outPorts: []*port.OutPort{port.NewOut(), port.NewOut()},
		errPort:  port.NewOut(),
	}

	n.inPort.AddInitHook(port.InitHookFunc(n.forward))
	n.outPorts[0].AddInitHook(port.InitHookFunc(n.rewrite))
	n.outPorts[1].AddInitHook(port.InitHookFunc(n.backward))
	n.errPort.AddInitHook(port.InitHookFunc(n.catch))

	return n
}

// In returns the input port with the specified name.
func (n *CallNode) In(name string) *port.InPort {
	n.mu.RLock()
	defer n.mu.RUnlock()

	switch name {
	case node.PortIn:
		return n.inPort
	default:
	}

	return nil
}

// Out returns the output port with the specified name.
func (n *CallNode) Out(name string) *port.OutPort {
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
func (n *CallNode) Close() error {
	n.mu.Lock()
	defer n.mu.Unlock()

	n.inPort.Close()
	for _, p := range n.outPorts {
		p.Close()
	}
	n.errPort.Close()
	n.tracers.Close()

	return nil
}

func (n *CallNode) forward(proc *process.Process) {
	n.mu.RLock()
	defer n.mu.RUnlock()

	tracer, _ := n.tracers.LoadOrStore(proc, func() (*packet.Tracer, error) {
		return packet.NewTracer(), nil
	})

	inReader := n.inPort.Open(proc)
	outWriter0 := n.outPorts[0].Open(proc)

	for {
		inPck, ok := <-inReader.Read()
		if !ok {
			return
		}
		tracer.Read(inReader, inPck)
		
		tracer.Write(outWriter0, inPck)
	}
}

func (n *CallNode) rewrite(proc *process.Process) {
	n.mu.RLock()
	defer n.mu.RUnlock()

	tracer, _ := n.tracers.LoadOrStore(proc, func() (*packet.Tracer, error) {
		return packet.NewTracer(), nil
	})

	outWriter0 := n.outPorts[0].Open(proc)
	outWriter1 := n.outPorts[1].Open(proc)
	errWriter := n.errPort.Open(proc)

	for {
		backPck, ok := <-outWriter0.Receive()
		if !ok {
			return
		}

		if _, ok := backPck.Payload().(object.Error); ok {
			tracer.Redirect(outWriter0, errWriter, backPck)
		} else {
			tracer.Redirect(outWriter0, outWriter1, backPck)
		}
	}
}

func (n *CallNode) backward(proc *process.Process) {
	n.mu.RLock()
	defer n.mu.RUnlock()

	tracer, _ := n.tracers.LoadOrStore(proc, func() (*packet.Tracer, error) {
		return packet.NewTracer(), nil
	})

	outWriter1 := n.outPorts[1].Open(proc)

	for {
		backPck, ok := <-outWriter1.Receive()
		if !ok {
			return
		}

		tracer.Receive(outWriter1, backPck)
	}
}

func (n *CallNode) catch(proc *process.Process) {
	n.mu.RLock()
	defer n.mu.RUnlock()

	tracer, _ := n.tracers.LoadOrStore(proc, func() (*packet.Tracer, error) {
		return packet.NewTracer(), nil
	})

	errWriter := n.errPort.Open(proc)

	for {
		backPck, ok := <-errWriter.Receive()
		if !ok {
			return
		}

		tracer.Receive(errWriter, backPck)
	}
}

// NewCallNodeCodec creates a new codec for CallNodeSpec.
func NewCallNodeCodec() scheme.Codec {
	return scheme.CodecWithType(func(spec *CallNodeSpec) (node.Node, error) {
		n := NewCallNode()

		return n, nil
	})
}
