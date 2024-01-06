package control

import (
	"github.com/samber/lo"
	"github.com/siyul-park/uniflow/pkg/node"
	"github.com/siyul-park/uniflow/pkg/packet"
	"github.com/siyul-park/uniflow/pkg/primitive"
	"github.com/siyul-park/uniflow/pkg/process"
)

type MergeNode struct {
	*node.ManyToOneNode
}

const (
	ModeConcat = "concat"
	ModeZip    = "zip"
)

var _ node.Node = (*MergeNode)(nil)

func NewMergeNode(mode string) *MergeNode {
	n := &MergeNode{}
	if mode == ModeConcat {
		n.ManyToOneNode = node.NewManyToOneNode(n.concat)
	} else {
		n.ManyToOneNode = node.NewManyToOneNode(n.zip)
	}
	return n
}

func (n *MergeNode) concat(proc *process.Process, inPcks []*packet.Packet) (*packet.Packet, *packet.Packet) {
	for _, inPck := range inPcks {
		if inPck == nil {
			return nil, nil
		}
	}

	inPayloads := lo.Map[*packet.Packet, primitive.Value](inPcks, func(item *packet.Packet, _ int) primitive.Value {
		return item.Payload()
	})
	outPayload := primitive.NewSlice(inPayloads...)

	return packet.New(outPayload), nil
}

func (n *MergeNode) zip(proc *process.Process, inPcks []*packet.Packet) (*packet.Packet, *packet.Packet) {
	for _, inPck := range inPcks {
		if inPck == nil {
			return nil, nil
		}
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
