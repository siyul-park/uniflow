package control

import (
	"time"

	"github.com/siyul-park/uniflow/pkg/node"
	"github.com/siyul-park/uniflow/pkg/packet"
	"github.com/siyul-park/uniflow/pkg/process"
	"github.com/siyul-park/uniflow/pkg/scheme"
	"github.com/siyul-park/uniflow/pkg/spec"
)

// SleepNodeSpec defines the configuration for SleepNode.
type SleepNodeSpec struct {
	spec.Meta `map:",inline"`
	Interval  time.Duration `map:"interval" validate:"required"`
}

// SleepNode adds a delay to packet processing, using a specified interval.
type SleepNode struct {
	*node.OneToOneNode
	interval time.Duration
}

const KindSleep = "sleep"

// NewSleepNodeCodec creates a codec to build SleepNode from SleepNodeSpec.
func NewSleepNodeCodec() scheme.Codec {
	return scheme.CodecWithType(func(spec *SleepNodeSpec) (node.Node, error) {
		return NewSleepNode(spec.Interval), nil
	})
}

// NewSleepNode creates a SleepNode with the given delay interval.
func NewSleepNode(interval time.Duration) *SleepNode {
	n := &SleepNode{interval: interval}
	n.OneToOneNode = node.NewOneToOneNode(n.action)
	return n
}

func (n *SleepNode) action(_ *process.Process, inPck *packet.Packet) (*packet.Packet, *packet.Packet) {
	time.Sleep(n.interval)
	return inPck, nil
}
