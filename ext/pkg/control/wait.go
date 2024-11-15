package control

import (
	"github.com/siyul-park/uniflow/pkg/node"
	"github.com/siyul-park/uniflow/pkg/packet"
	"github.com/siyul-park/uniflow/pkg/process"
	"github.com/siyul-park/uniflow/pkg/scheme"
	"github.com/siyul-park/uniflow/pkg/spec"
	"time"
)

// WaitNodeSpec defines the configuration for WaitNode.
type WaitNodeSpec struct {
	spec.Meta `map:",inline"`
	Interval  time.Duration `map:"interval"`
}

// WaitNode adds a delay to packet processing, using a specified interval.
type WaitNode struct {
	*node.OneToOneNode
	interval time.Duration
}

const KindWait = "wait"

// NewWaitNodeCodec creates a codec to build WaitNode from WaitNodeSpec.
func NewWaitNodeCodec() scheme.Codec {
	return scheme.CodecWithType(func(spec *WaitNodeSpec) (node.Node, error) {
		return NewWaitNode(spec.Interval), nil
	})
}

// NewWaitNode creates a WaitNode with the given delay interval.
func NewWaitNode(interval time.Duration) *WaitNode {
	n := &WaitNode{interval: interval}
	n.OneToOneNode = node.NewOneToOneNode(n.action)
	return n
}

func (n *WaitNode) action(_ *process.Process, inPck *packet.Packet) (*packet.Packet, *packet.Packet) {
	time.Sleep(n.interval)
	return inPck, nil
}
