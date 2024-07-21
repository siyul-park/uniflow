package control

import (
	"github.com/siyul-park/uniflow/pkg/node"
	"github.com/siyul-park/uniflow/pkg/packet"
	"github.com/siyul-park/uniflow/pkg/process"
	"github.com/siyul-park/uniflow/pkg/scheme"
	"github.com/siyul-park/uniflow/pkg/spec"
	"github.com/siyul-park/uniflow/pkg/types"
)

// SplitNodeSpec defines the specifications for creating a SpliteNode.
type SplitNodeSpec struct {
	spec.Meta `map:",inline"`
}

// SplitNode represents a node that splits incoming packets into multiple packets.
type SplitNode struct {
	*node.OneToManyNode
}

const KindSplit = "split"

// NewSplitNodeCodec creates and returns a codec for decoding SpliteNodeSpec.
func NewSplitNodeCodec() scheme.Codec {
	return scheme.CodecWithType(func(_ *SplitNodeSpec) (node.Node, error) {
		return NewSplitNode(), nil
	})
}

// NewSplitNode initializes and returns a new instance of SpliteNode.
func NewSplitNode() *SplitNode {
	n := &SplitNode{}
	n.OneToManyNode = node.NewOneToManyNode(n.action)
	return n
}

func (n *SplitNode) action(_ *process.Process, inPck *packet.Packet) ([]*packet.Packet, *packet.Packet) {
	switch inPayload := inPck.Payload().(type) {
	case types.Slice:
		outPcks := make([]*packet.Packet, 0, inPayload.Len())
		for i := 0; i < inPayload.Len(); i++ {
			outPck := packet.New(inPayload.Get(i))
			outPcks = append(outPcks, outPck)
		}
		return outPcks, nil
	default:
		return []*packet.Packet{packet.New(inPayload)}, nil
	}
}
