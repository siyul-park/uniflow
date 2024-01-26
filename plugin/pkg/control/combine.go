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
	mode string
	mu   sync.RWMutex
}

// CombineNodeSpec holds the specifications for creating a CombineNode.
type CombineNodeSpec struct {
	scheme.SpecMeta
	Mode string `map:"mode"`
}

const KindCombine = "combine"

const (
	ModeConcat = "concat"
	ModeZip    = "zip"
)

var _ node.Node = (*CombineNode)(nil)

// NewCombineNodeCodec creates a new codec for CombineNodeSpec.
func NewCombineNodeCodec() scheme.Codec {
	return scheme.CodecWithType[*CombineNodeSpec](func(spec *CombineNodeSpec) (node.Node, error) {
		return NewCombineNode(spec.Mode), nil
	})
}

// NewCombineNode creates a new CombineNode with the specified mode.
func NewCombineNode(mode string) *CombineNode {
	n := &CombineNode{mode: mode}
	n.ManyToOneNode = node.NewManyToOneNode(n.action)
	return n
}

func (n *CombineNode) action(proc *process.Process, inPcks []*packet.Packet) (*packet.Packet, *packet.Packet) {
	n.mu.RLock()
	defer n.mu.RUnlock()

	if n.mode == ModeConcat {
		return n.concat(proc, inPcks)
	} else {
		return n.zip(proc, inPcks)
	}
}

func (n *CombineNode) concat(proc *process.Process, inPcks []*packet.Packet) (*packet.Packet, *packet.Packet) {
	if !n.isFull(inPcks) {
		return nil, nil
	}

	inPayloads := lo.Map[*packet.Packet, primitive.Value](inPcks, func(item *packet.Packet, _ int) primitive.Value {
		return item.Payload()
	})
	outPayload := primitive.NewSlice(inPayloads...)

	return packet.New(outPayload), nil
}

func (n *CombineNode) zip(proc *process.Process, inPcks []*packet.Packet) (*packet.Packet, *packet.Packet) {
	if !n.isFull(inPcks) {
		return nil, nil
	}

	inPayloads := lo.Map[*packet.Packet, primitive.Value](inPcks, func(item *packet.Packet, _ int) primitive.Value {
		return item.Payload()
	})
	outPayload := lo.Reduce[primitive.Value, primitive.Value](inPayloads, func(pre, cur primitive.Value, index int) primitive.Value {
		switch pre := pre.(type) {
		case *primitive.Map:
			if cur, ok := cur.(*primitive.Map); ok {
				return primitive.NewMap(append(pre.Pairs(), cur.Pairs()...)...)
			}
		case *primitive.Slice:
			if cur, ok := cur.(*primitive.Slice); ok {
				return primitive.NewSlice(append(pre.Values(), cur.Values()...)...)
			}
			return primitive.NewSlice(append(pre.Values(), cur)...)
		}

		return cur
	}, nil)

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
