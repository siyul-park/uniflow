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
	depth int
	mu    sync.RWMutex
}

// CombineNodeSpec holds the specifications for creating a CombineNode.
type CombineNodeSpec struct {
	scheme.SpecMeta
	Depth int `map:"depth"`
}

const KindCombine = "combine"

var _ node.Node = (*CombineNode)(nil)

// NewCombineNodeCodec creates a new codec for CombineNodeSpec.
func NewCombineNodeCodec() scheme.Codec {
	return scheme.CodecWithType[*CombineNodeSpec](func(spec *CombineNodeSpec) (node.Node, error) {
		return NewCombineNode(spec.Depth), nil
	})
}

// NewCombineNode creates a new CombineNode.
func NewCombineNode(depth int) *CombineNode {
	n := &CombineNode{depth: depth}
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
	return y
}
