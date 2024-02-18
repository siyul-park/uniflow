package control

import (
	"sync"

	"github.com/samber/lo"
	"github.com/siyul-park/uniflow/pkg/node"
	"github.com/siyul-park/uniflow/pkg/packet"
	"github.com/siyul-park/uniflow/pkg/primitive"
	"github.com/siyul-park/uniflow/pkg/process"
	"github.com/siyul-park/uniflow/pkg/scheme"
)

// CombineNode represents a node that Combines multiple input packets into a single output packet.
type CombineNode struct {
	*node.ManyToOneNode
	depth   int
	inplace bool
	mu      sync.RWMutex
}

// CombineNodeSpec holds the specifications for creating a CombineNode.
type CombineNodeSpec struct {
	scheme.SpecMeta `map:",inline"`
	Depth           int  `map:"depth"`
	Inplace         bool `map:"inplace"`
}

const KindCombine = "combine"

// NewCombineNode creates a new CombineNode.
func NewCombineNode(depth int, inplace bool) *CombineNode {
	n := &CombineNode{depth: depth, inplace: inplace}
	n.ManyToOneNode = node.NewManyToOneNode(n.action)
	return n
}

func (n *CombineNode) action(proc *process.Process, inPcks []*packet.Packet) (*packet.Packet, *packet.Packet) {
	n.mu.RLock()
	defer n.mu.RUnlock()

	if !n.isFull(inPcks) {
		return nil, nil
	}

	inPayloads := lo.Map[*packet.Packet, primitive.Value](inPcks, func(item *packet.Packet, _ int) primitive.Value {
		return item.Payload()
	})

	var outPayload primitive.Value
	if n.depth == 0 {
		outPayload = primitive.NewSlice(inPayloads...)
	} else {
		outPayload = lo.Reduce[primitive.Value, primitive.Value](inPayloads, func(agg primitive.Value, item primitive.Value, index int) primitive.Value {
			return n.merge(agg, item, n.depth-1)
		}, nil)
	}

	return packet.New(outPayload), nil
}

func (n *CombineNode) isFull(pcks []*packet.Packet) bool {
	for _, inPck := range pcks {
		if inPck == nil {
			return false
		}
	}
	return true
}

func (n *CombineNode) merge(x, y primitive.Value, depth int) primitive.Value {
	if x == nil {
		return y
	}
	if y == nil {
		return x
	}

	if depth == 0 {
		return y
	}

	switch x := x.(type) {
	case *primitive.Slice:
		if y, ok := y.(*primitive.Slice); ok {
			var values []primitive.Value
			if n.inplace {
				len := x.Len()
				if len < y.Len() {
					len = y.Len()
				}
				for i := 0; i < len; i++ {
					value1 := x.Get(i)
					value2 := y.Get(i)

					values = append(values, n.merge(value1, value2, depth-1))
				}
			} else {
				values = append(x.Values(), y.Values()...)
			}

			return primitive.NewSlice(values...)
		}
	case *primitive.Map:
		if y, ok := y.(*primitive.Map); ok {
			var pairs []primitive.Value
			for _, key := range x.Keys() {
				value1, _ := x.Get(key)
				value2, _ := y.Get(key)

				pairs = append(pairs, key, n.merge(value1, value2, depth-1))
			}
			for _, key := range y.Keys() {
				_, ok := x.Get(key)
				value, _ := y.Get(key)
				if ok {
					continue
				}
				pairs = append(pairs, key, value)
			}

			return primitive.NewMap(pairs...)
		}
	}

	return y
}

// NewCombineNodeCodec creates a new codec for CombineNodeSpec.
func NewCombineNodeCodec() scheme.Codec {
	return scheme.CodecWithType[*CombineNodeSpec](func(spec *CombineNodeSpec) (node.Node, error) {
		return NewCombineNode(spec.Depth, spec.Inplace), nil
	})
}
