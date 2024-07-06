package control

import (
	"sync"

	"github.com/siyul-park/uniflow/pkg/node"
	"github.com/siyul-park/uniflow/pkg/port"
	"github.com/siyul-park/uniflow/pkg/process"
	"github.com/siyul-park/uniflow/pkg/scheme"
	"github.com/siyul-park/uniflow/pkg/spec"
	"github.com/siyul-park/uniflow/pkg/symbol"
)

// BlockNode is a node that handles multiple sub-nodes.
type BlockNode struct {
	nodes    []node.Node
	inPorts  map[string]*port.InPort
	outPorts map[string]*port.OutPort
	mu       sync.RWMutex
}

// BlockNodeSpec defines the specification for creating a BlockNode.
type BlockNodeSpec struct {
	spec.Meta `map:",inline"`
	Specs     []*spec.Unstructured `map:"specs"`
}

const KindBlock = "block"

var _ node.Node = (*BlockNode)(nil)

// BlockNodeSpec defines the specification for creating a BlockNode.
func NewBlockNode(nodes ...node.Node) *BlockNode {
	n := &BlockNode{
		nodes:    nodes,
		inPorts:  make(map[string]*port.InPort),
		outPorts: make(map[string]*port.OutPort),
	}

	for i := 1; i < len(n.nodes); i++ {
		out := n.nodes[i-1].Out(node.PortOut)
		in := n.nodes[i].In(node.PortIn)
		if out == nil || in == nil {
			break
		}
		out.Link(in)
	}

	if len(n.nodes) > 1 {
		inPort := port.NewIn()
		outPort := port.NewOut()

		inPort.Accept(port.ListenFunc(n.throw))
		outPort.Accept(port.ListenFunc(n.catch))

		n.inPorts[node.PortErr] = inPort
		n.outPorts[node.PortErr] = outPort

		for _, cur := range n.nodes {
			if err := cur.Out(node.PortErr); err != nil {
				err.Link(inPort)
			}
		}
	}

	return n
}

// Nodes returns the sub-nodes.
func (n *BlockNode) Nodes() []node.Node {
	n.mu.RLock()
	defer n.mu.RUnlock()

	return n.nodes
}

// In returns the input port with the specified name.
func (n *BlockNode) In(name string) *port.InPort {
	n.mu.Lock()
	defer n.mu.Unlock()

	if p, ok := n.inPorts[name]; ok {
		return p
	}
	if len(n.nodes) > 0 {
		if p := n.nodes[0].In(name); p != nil {
			n.inPorts[name] = p
			return p
		}
	}
	return nil
}

// Out returns the output port with the specified name.
func (n *BlockNode) Out(name string) *port.OutPort {
	n.mu.Lock()
	defer n.mu.Unlock()

	if p, ok := n.outPorts[name]; ok {
		return p
	}
	if len(n.nodes) > 0 {
		if p := n.nodes[len(n.nodes)-1].Out(name); p != nil {
			n.outPorts[name] = p
			return p
		}
	}
	return nil
}

// Close closes all ports associated with the node.
func (n *BlockNode) Close() error {
	n.mu.Lock()
	defer n.mu.Unlock()

	for _, n := range n.nodes {
		if err := n.Close(); err != nil {
			return err
		}
	}

	for name, p := range n.inPorts {
		p.Close()
		delete(n.inPorts, name)
	}
	for name, p := range n.outPorts {
		p.Close()
		delete(n.outPorts, name)
	}

	return nil
}

func (n *BlockNode) throw(proc *process.Process) {
	n.mu.RLock()
	defer n.mu.RUnlock()

	errWriter := n.outPorts[node.PortErr].Open(proc)
	errReader := n.inPorts[node.PortErr].Open(proc)

	for {
		inPck, ok := <-errReader.Read()
		if !ok {
			return
		}

		if errWriter.Write(inPck) == 0 {
			errReader.Receive(inPck)
		}
	}
}

func (n *BlockNode) catch(proc *process.Process) {
	n.mu.RLock()
	defer n.mu.RUnlock()

	errWriter := n.outPorts[node.PortErr].Open(proc)
	errReader := n.inPorts[node.PortErr].Open(proc)

	for {
		backPck, ok := <-errWriter.Receive()
		if !ok {
			return
		}

		errReader.Receive(backPck)
	}
}

// NewBlockNodeCodec creates a new codec for IfNodeSpec.
func NewBlockNodeCodec(s *scheme.Scheme) scheme.Codec {
	return scheme.CodecWithType(func(spec *BlockNodeSpec) (node.Node, error) {
		nodes := make([]node.Node, 0, len(spec.Specs))
		for _, spec := range spec.Specs {
			n, err := s.Decode(spec)
			if err != nil {
				for _, n := range nodes {
					n.Close()
				}
				return nil, err
			}
			nodes = append(nodes, symbol.New(spec, n))
		}
		return NewBlockNode(nodes...), nil
	})
}
