package control

import (
	"sync"

	"github.com/siyul-park/uniflow/pkg/node"
	"github.com/siyul-park/uniflow/pkg/port"
	"github.com/siyul-park/uniflow/pkg/process"
	"github.com/siyul-park/uniflow/pkg/scheme"
	"github.com/siyul-park/uniflow/pkg/spec"
)

// BlockNode is a node that handles multiple sub-nodes.
type BlockNode struct {
	nodes    []node.Node
	eoutPort *port.OutPort
	einPort  *port.InPort
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
		eoutPort: port.NewOut(),
		einPort:  port.NewIn(),
	}

	n.eoutPort.AddInitHook(port.InitHookFunc(n.backward))
	n.einPort.AddInitHook(port.InitHookFunc(n.forward))

	for i := 1; i < len(n.nodes); i++ {
		out := n.nodes[i-1].Out(node.PortOut)
		in := n.nodes[i].In(node.PortIn)
		if out == nil || in == nil {
			break
		}
		out.Link(in)
	}

	for _, cur := range n.nodes {
		if err := cur.Out(node.PortErr); err != nil {
			err.Link(n.einPort)
		}
	}

	return n
}

// In returns the input port with the specified name.
func (n *BlockNode) In(name string) *port.InPort {
	n.mu.RLock()
	defer n.mu.RUnlock()

	if len(n.nodes) > 0 {
		return n.nodes[0].In(name)
	}
	return nil
}

// Out returns the output port with the specified name.
func (n *BlockNode) Out(name string) *port.OutPort {
	n.mu.RLock()
	defer n.mu.RUnlock()

	switch name {
	case node.PortErr:
		return n.eoutPort
	default:
		if len(n.nodes) > 0 {
			return n.nodes[len(n.nodes)-1].Out(name)
		}
	}

	return nil
}

// Close closes all ports associated with the node.
func (n *BlockNode) Close() error {
	n.mu.Lock()
	defer n.mu.Unlock()

	n.eoutPort.Close()
	n.einPort.Close()

	for _, n := range n.nodes {
		if err := n.Close(); err != nil {
			return err
		}
	}

	return nil
}

func (n *BlockNode) forward(proc *process.Process) {
	n.mu.RLock()
	defer n.mu.RUnlock()

	errWriter := n.eoutPort.Open(proc)
	errReader := n.einPort.Open(proc)

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

func (n *BlockNode) backward(proc *process.Process) {
	n.mu.RLock()
	defer n.mu.RUnlock()

	errWriter := n.eoutPort.Open(proc)
	errReader := n.einPort.Open(proc)

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
			nodes = append(nodes, n)
		}
		return NewBlockNode(nodes...), nil
	})
}
