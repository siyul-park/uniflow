package control

import (
	"sync"

	"github.com/siyul-park/uniflow/pkg/node"
	"github.com/siyul-park/uniflow/pkg/packet"
	"github.com/siyul-park/uniflow/pkg/port"
	"github.com/siyul-park/uniflow/pkg/process"
	"github.com/siyul-park/uniflow/pkg/scheme"
)

// GotoNode redirects packets from the input port to the intermediate port for processing by connected nodes, then outputs the results to the output port.
type GotoNode struct {
	inPort   *port.InPort
	outPorts []*port.OutPort
	errPort  *port.OutPort
	mu       sync.RWMutex
}

// GotoNodeSpec holds the specifications for creating a GotoNode.
type GotoNodeSpec struct {
	scheme.SpecMeta `map:",inline"`
}

var _ node.Node = (*GotoNode)(nil)

const KindGoto = "Goto"

// NewGotoNode creates a new GotoNode.
func NewGotoNode() *GotoNode {
	n := &GotoNode{
		inPort:   port.NewIn(),
		outPorts: []*port.OutPort{port.NewOut(), port.NewOut()},
		errPort:  port.NewOut(),
	}

	n.inPort.AddHandler(port.HandlerFunc(n.forward))
	n.outPorts[0].AddHandler(port.HandlerFunc(n.redirect))
	n.outPorts[1].AddHandler(port.HandlerFunc(n.backward))
	n.errPort.AddHandler(port.HandlerFunc(n.catch))

	return n
}

// In returns the input port with the specified name.
func (n *GotoNode) In(name string) *port.InPort {
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
func (n *GotoNode) Out(name string) *port.OutPort {
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
func (n *GotoNode) Close() error {
	n.mu.Lock()
	defer n.mu.Unlock()

	n.inPort.Close()
	for _, p := range n.outPorts {
		p.Close()
	}
	n.errPort.Close()

	return nil
}

func (n *GotoNode) forward(proc *process.Process) {
	n.mu.RLock()
	defer n.mu.RUnlock()

	inReader := n.inPort.Open(proc)
	outWriter0 := n.outPorts[0].Open(proc)

	for {
		inPck, ok := <-inReader.Read()
		if !ok {
			return
		}

		outWriter0.Write(inPck)
	}
}

func (n *GotoNode) redirect(proc *process.Process) {
	n.mu.RLock()
	defer n.mu.RUnlock()

	outWriter0 := n.outPorts[0].Open(proc)
	outWriter1 := n.outPorts[1].Open(proc)

	for {
		backPck, ok := <-outWriter0.Receive()
		if !ok {
			return
		}

		if _, ok := packet.AsError(backPck); ok {
			n.throw(proc, backPck)
		} else {
			outWriter1.Write(backPck)
		}
	}
}

func (n *GotoNode) backward(proc *process.Process) {
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

func (n *GotoNode) catch(proc *process.Process) {
	n.mu.RLock()
	defer n.mu.RUnlock()

	inReader := n.inPort.Open(proc)
	errWriter := n.errPort.Open(proc)

	for {
		backPck, ok := <-errWriter.Receive()
		if !ok {
			return
		}

		inReader.Receive(backPck)
	}
}

func (n *GotoNode) throw(proc *process.Process, errPck *packet.Packet) {
	inReader := n.inPort.Open(proc)
	errWriter := n.errPort.Open(proc)

	if errWriter.Links() > 0 {
		errWriter.Write(errPck)
	} else {
		inReader.Receive(errPck)
	}
}

// NewGotoNodeCodec creates a new codec for GotoNodeSpec.
func NewGotoNodeCodec() scheme.Codec {
	return scheme.CodecWithType[*GotoNodeSpec](func(spec *GotoNodeSpec) (node.Node, error) {
		return NewGotoNode(), nil
	})
}
