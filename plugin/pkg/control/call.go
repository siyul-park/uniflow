package control

import (
	"sync"

	"github.com/siyul-park/uniflow/pkg/node"
	"github.com/siyul-park/uniflow/pkg/packet"
	"github.com/siyul-park/uniflow/pkg/port"
	"github.com/siyul-park/uniflow/pkg/process"
	"github.com/siyul-park/uniflow/pkg/scheme"
)

// CallNode redirects packets from the input port to the intermediate port for processing by connected nodes, then outputs the results to the output port.
type CallNode struct {
	bridges  *process.Local[*packet.Bridge]
	inPort   *port.InPort
	outPorts []*port.OutPort
	errPort  *port.OutPort
	mu       sync.RWMutex
}

// CallNodeSpec holds the specifications for creating a CallNode.
type CallNodeSpec struct {
	scheme.SpecMeta `map:",inline"`
}

var _ node.Node = (*CallNode)(nil)

const KindCall = "call"

// NewCallNode creates a new CallNode.
func NewCallNode() *CallNode {
	n := &CallNode{
		bridges:  process.NewLocal[*packet.Bridge](),
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

	return nil
}

func (n *CallNode) forward(proc *process.Process) {
	n.mu.RLock()
	defer n.mu.RUnlock()

	bridge, _ := n.bridges.LoadOrStore(proc, func() (*packet.Bridge, error) {
		return packet.NewBridge(), nil
	})

	inReader := n.inPort.Open(proc)
	outWriter0 := n.outPorts[0].Open(proc)

	for {
		inPck, ok := <-inReader.Read()
		if !ok {
			return
		}

		bridge.Write([]*packet.Packet{inPck}, []*packet.Reader{inReader}, []*packet.Writer{outWriter0})
	}
}

func (n *CallNode) rewrite(proc *process.Process) {
	n.mu.RLock()
	defer n.mu.RUnlock()

	bridge, _ := n.bridges.LoadOrStore(proc, func() (*packet.Bridge, error) {
		return packet.NewBridge(), nil
	})

	outWriter0 := n.outPorts[0].Open(proc)
	outWriter1 := n.outPorts[1].Open(proc)
	errWriter := n.errPort.Open(proc)

	for {
		backPck, ok := <-outWriter0.Receive()
		if !ok {
			return
		}

		if _, ok := packet.AsError(backPck); ok {
			bridge.Rewrite(backPck, outWriter0, errWriter)
		} else {
			bridge.Rewrite(backPck, outWriter0, outWriter1)
		}
	}
}

func (n *CallNode) backward(proc *process.Process) {
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

func (n *CallNode) catch(proc *process.Process) {
	n.mu.RLock()
	defer n.mu.RUnlock()

	bridge, _ := n.bridges.LoadOrStore(proc, func() (*packet.Bridge, error) {
		return packet.NewBridge(), nil
	})

	errWriter := n.errPort.Open(proc)

	for {
		backPck, ok := <-errWriter.Receive()
		if !ok {
			return
		}

		bridge.Receive(backPck, errWriter)
	}
}

// NewCallNodeCodec creates a new codec for CallNodeSpec.
func NewCallNodeCodec() scheme.Codec {
	return scheme.CodecWithType(func(spec *CallNodeSpec) (node.Node, error) {
		n := NewCallNode()

		return n, nil
	})
}
