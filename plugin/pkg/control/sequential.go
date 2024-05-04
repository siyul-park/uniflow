package control

import (
	"sync"

	"github.com/siyul-park/uniflow/pkg/node"
	"github.com/siyul-park/uniflow/pkg/port"
	"github.com/siyul-park/uniflow/pkg/scheme"
)

// SequentialNode represents a node that processes its children sequentially.
type SequentialNode struct {
	children []node.Node
	mu       sync.RWMutex
}

// SequentialNodeSpec contains specifications for creating a SequentialNode.
type SequentialNodeSpec struct {
	scheme.SpecMeta `map:",inline"`
	Children        []*scheme.Unstructured `map:",children"`
}

var _ node.Node = (*SequentialNode)(nil)

const KindSequential = "sequential"

// NewSequentialNode creates a new SequentialNode with the provided children.
func NewSequentialNode(children ...node.Node) *SequentialNode {
	for i, cur := range children {
		if i > 0 {
			pre := children[i-1]

			out := pre.Out(node.PortOut)
			in := cur.In(node.PortIn)

			out.Link(in)
		}
	}

	return &SequentialNode{
		children: children,
	}
}

// Children returns the children nodes of the SequentialNode.
func (n *SequentialNode) Children() []node.Node {
	n.mu.RLock()
	defer n.mu.RUnlock()

	return n.children[:]
}

// In returns the input port with the specified name.
func (n *SequentialNode) In(name string) *port.InPort {
	n.mu.RLock()
	defer n.mu.RUnlock()

	var child node.Node
	if len(n.children) > 0 {
		child = n.children[0]
	}

	if child != nil {
		return child.In(name)
	}
	return nil
}

// Out returns the output port with the specified name.
func (n *SequentialNode) Out(name string) *port.OutPort {
	n.mu.RLock()
	defer n.mu.RUnlock()

	var child node.Node
	if len(n.children) > 0 {
		if name == node.PortErr {
			child = n.children[0]
		} else {
			child = n.children[len(n.children)-1]
		}
	}

	if child != nil {
		return child.Out(name)
	}
	return nil
}

// Close closes all ports associated with the SequentialNode.
func (n *SequentialNode) Close() error {
	n.mu.Lock()
	defer n.mu.Unlock()

	for _, child := range n.children {
		if err := child.Close(); err != nil {
			return err
		}
	}

	return nil
}

// NewSequentialNodeCodec creates a scheme codec for SequentialNode.
func NewSequentialNodeCodec(s *scheme.Scheme) scheme.Codec {
	return scheme.CodecWithType(func(spec *SequentialNodeSpec) (node.Node, error) {
		var children []node.Node
		for _, child := range spec.Children {
			if n, err := s.Decode(child); err != nil {
				for _, c := range children {
					_ = c.Close()
				}
				return nil, err
			} else {
				children = append(children, n)
			}
		}
		return NewSequentialNode(children...), nil
	})
}
