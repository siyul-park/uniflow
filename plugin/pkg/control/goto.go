package control

import (
	"sync"

	"github.com/siyul-park/uniflow/pkg/node"
	"github.com/siyul-park/uniflow/pkg/packet"
	"github.com/siyul-park/uniflow/pkg/port"
	"github.com/siyul-park/uniflow/pkg/process"
	"github.com/siyul-park/uniflow/pkg/scheme"
)

// GoToNode redirects packets from the input port to the intermediate port for processing by connected nodes, then outputs the results to the output port.
type GoToNode struct {
	inPort   *port.InPort
	outPorts []*port.OutPort
	errPort  *port.OutPort
	mu       sync.RWMutex
}

// GoToNodeSpec holds the specifications for creating a GoToNode.
type GoToNodeSpec struct {
	scheme.SpecMeta `map:",inline"`
}

var _ node.Node = (*GoToNode)(nil)

const KindGoTo = "goto"

// NewGoToNode creates a new GoToNode.
func NewGoToNode() *GoToNode {
	n := &GoToNode{
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
func (n *GoToNode) In(name string) *port.InPort {
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
func (n *GoToNode) Out(name string) *port.OutPort {
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
func (n *GoToNode) Close() error {
	n.mu.Lock()
	defer n.mu.Unlock()

	n.inPort.Close()
	for _, p := range n.outPorts {
		p.Close()
	}
	n.errPort.Close()

	return nil
}

func (n *GoToNode) forward(proc *process.Process) {
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

func (n *GoToNode) redirect(proc *process.Process) {
	n.mu.RLock()
	defer n.mu.RUnlock()

	inReader := n.inPort.Open(proc)
	outWriter0 := n.outPorts[0].Open(proc)
	outWriter1 := n.outPorts[1].Open(proc)

	for {
		backPck, ok := <-outWriter0.Receive()
		if !ok {
			return
		}

		if _, ok := packet.AsError(backPck); ok {
			n.throw(proc, backPck)
		} else if outWriter1.Links() > 0 {
			outWriter1.Write(backPck)
		} else {
			inReader.Receive(backPck)
		}
	}
}

func (n *GoToNode) backward(proc *process.Process) {
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

func (n *GoToNode) catch(proc *process.Process) {
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

func (n *GoToNode) throw(proc *process.Process, errPck *packet.Packet) {
	inReader := n.inPort.Open(proc)
	errWriter := n.errPort.Open(proc)

	if errWriter.Links() > 0 {
		errWriter.Write(errPck)
	} else {
		inReader.Receive(errPck)
	}
}

// NewGoToNodeCodec creates a new codec for GoToNodeSpec.
func NewGoToNodeCodec() scheme.Codec {
	return scheme.CodecWithType(func(spec *GoToNodeSpec) (node.Node, error) {
		return NewGoToNode(), nil
	})
}
