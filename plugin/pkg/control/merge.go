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

// MergeNode represents a node that merges multiple input packets into a single output packet.
type MergeNode struct {
	*node.ManyToOneNode
	mode string
	mu   sync.RWMutex
}

// MergeNodeSpec holds the specifications for creating a MergeNode.
type MergeNodeSpec struct {
	scheme.SpecMeta
	Mode string `map:"mode"`
}

const KindMerge = "merge"

const (
	ModeConcat = "concat"
	ModeZip    = "zip"
)

var _ node.Node = (*MergeNode)(nil)

// NewMergeNodeCodec creates a new codec for MergeNodeSpec.
func NewMergeNodeCodec() scheme.Codec {
	return scheme.CodecWithType[*MergeNodeSpec](func(spec *MergeNodeSpec) (node.Node, error) {
		return NewMergeNode(spec.Mode), nil
	})
}

// NewMergeNode creates a new MergeNode with the specified mode.
func NewMergeNode(mode string) *MergeNode {
	n := &MergeNode{mode: mode}
	n.ManyToOneNode = node.NewManyToOneNode(n.action)
	return n
}

func (n *MergeNode) action(proc *process.Process, inPcks []*packet.Packet) (*packet.Packet, *packet.Packet) {
	n.mu.RLock()
	defer n.mu.RUnlock()

	if n.mode == ModeConcat {
		return n.concat(proc, inPcks)
	} else {
		return n.zip(proc, inPcks)
	}
}

func (n *MergeNode) concat(proc *process.Process, inPcks []*packet.Packet) (*packet.Packet, *packet.Packet) {
	if !n.isFull(inPcks) {
		return nil, nil
	}

	inPayloads := lo.Map[*packet.Packet, primitive.Value](inPcks, func(item *packet.Packet, _ int) primitive.Value {
		return item.Payload()
	})
	outPayload := primitive.NewSlice(inPayloads...)

	return packet.New(outPayload), nil
}

func (n *MergeNode) zip(proc *process.Process, inPcks []*packet.Packet) (*packet.Packet, *packet.Packet) {
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

func (n *MergeNode) isFull(pcks []*packet.Packet) bool {
	for _, inPck := range pcks {
		if inPck == nil {
			return false
		}
	}
	return true
}
