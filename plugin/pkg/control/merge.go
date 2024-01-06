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

var _ node.Node = (*MergeNode)(nil)

func NewMergeNode() *MergeNode {
	n := &MergeNode{}
	n.ManyToOneNode = node.NewManyToOneNode(n.action)
	return n
}

func (n *MergeNode) action(proc *process.Process, inPcks []*packet.Packet) (*packet.Packet, *packet.Packet) {
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
