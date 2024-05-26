package control

import (
	"sync"

	"github.com/siyul-park/uniflow/pkg/node"
	"github.com/siyul-park/uniflow/pkg/packet"
	"github.com/siyul-park/uniflow/pkg/process"
	"github.com/siyul-park/uniflow/pkg/scheme"
)

// MergeNode represents a node that Merges multiple input packets into a single output packet.
type MergeNode struct {
	*node.ManyToOneNode
	mu sync.RWMutex
}

// MergeNodeSpec holds the specifications for creating a MergeNode.
type MergeNodeSpec struct {
	scheme.SpecMeta `map:",inline"`
}

const KindMerge = "merge"

// NewMergeNode creates a new MergeNode.
func NewMergeNode() *MergeNode {
	n := &MergeNode{}

	n.ManyToOneNode = node.NewManyToOneNode(n.action)

	return n
}

func (n *MergeNode) action(proc *process.Process, inPcks []*packet.Packet) (*packet.Packet, *packet.Packet) {
	n.mu.RLock()
	defer n.mu.RUnlock()

	for _, inPck := range inPcks {
		if inPck == nil {
			return nil, nil
		}
	}

	return packet.Merge(inPcks), nil
}

// NewMergeNodeCodec creates a new codec for MergeNodeSpec.
func NewMergeNodeCodec() scheme.Codec {
	return scheme.CodecWithType(func(spec *MergeNodeSpec) (node.Node, error) {
		return NewMergeNode(), nil
	})
}
