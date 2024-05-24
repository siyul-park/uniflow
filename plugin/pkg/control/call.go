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
		inPort:   port.NewIn(),
		outPorts: []*port.OutPort{port.NewOut(), port.NewOut()},
		errPort:  port.NewOut(),
	}

	n.inPort.AddInitHook(port.InitHookFunc(n.forward))

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

	inReader := n.inPort.Open(proc)
	outWriter0 := n.outPorts[0].Open(proc)
	outWriter1 := n.outPorts[0].Open(proc)
	errWriter := n.errPort.Open(proc)

	for {
		inPck, ok := <-inReader.Read()
		if !ok {
			return
		}

		backPck := port.Call(outWriter0, inPck)
		if backPck == packet.None {
			inReader.Receive(backPck)
			continue
		}

		if _, ok := packet.AsError(backPck); ok {
			backPck = port.Call(errWriter, backPck)
		} else {
			backPck = port.Call(outWriter1, backPck)
		}

		if backPck == packet.None {
			inReader.Receive(backPck)
		}
	}
}

// NewCallNodeCodec creates a new codec for CallNodeSpec.
func NewCallNodeCodec() scheme.Codec {
	return scheme.CodecWithType(func(spec *CallNodeSpec) (node.Node, error) {
		n := NewCallNode()

		return n, nil
	})
}
