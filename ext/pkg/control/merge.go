package control

import (
	"github.com/siyul-park/uniflow/pkg/node"
	"github.com/siyul-park/uniflow/pkg/packet"
	"github.com/siyul-park/uniflow/pkg/process"
	"github.com/siyul-park/uniflow/pkg/scheme"
	"github.com/siyul-park/uniflow/pkg/spec"
	"github.com/siyul-park/uniflow/pkg/types"
)


// MergeNodeSpec holds the specifications for creating a MergeNode.
type MergeNodeSpec struct {
	spec.Meta `map:",inline"`
}

// MergeNode represents a node that Merges multiple input packets into a single output packet.
type MergeNode struct {
	*node.ManyToOneNode
}

const KindMerge = "merge"

// NewMergeNodeCodec creates a new codec for MergeNodeSpec.
func NewMergeNodeCodec() scheme.Codec {
	return scheme.CodecWithType(func(spec *MergeNodeSpec) (node.Node, error) {
		return NewMergeNode(), nil
	})
}

// NewMergeNode creates a new MergeNode.
func NewMergeNode() *MergeNode {
	n := &MergeNode{}
	n.ManyToOneNode = node.NewManyToOneNode(n.action)
	return n
}

func (n *MergeNode) action(proc *process.Process, inPcks []*packet.Packet) (*packet.Packet, *packet.Packet) {
	outPck := packet.Merge(inPcks)

	if _, ok := outPck.Payload().(types.Error); ok {
		return nil, outPck
	} else {
		return outPck, nil
	}
}

