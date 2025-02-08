package control

import (
	"github.com/siyul-park/uniflow/pkg/node"
	"github.com/siyul-park/uniflow/pkg/packet"
	"github.com/siyul-park/uniflow/pkg/port"
	"github.com/siyul-park/uniflow/pkg/process"
	"github.com/siyul-park/uniflow/pkg/scheme"
	"github.com/siyul-park/uniflow/pkg/spec"
)

// NOPNodeSpec defines the specification for creating a NOPNode.
type NOPNodeSpec struct {
	spec.Meta `json:",inline"`
}

// NOPNode represents a node that performs no operation and simply forwards incoming packets.
type NOPNode struct {
	inPort *port.InPort
}

const KindNOP = "nop"

var _ node.Node = (*NOPNode)(nil)

// NewNOPNodeCodec creates a codec for decoding NOPNodeSpec.
func NewNOPNodeCodec() scheme.Codec {
	return scheme.CodecWithType(func(_ *NOPNodeSpec) (node.Node, error) {
		return NewNOPNode(), nil
	})
}

// NewNOPNode creates a new instance of NOPNode.
func NewNOPNode() *NOPNode {
	n := &NOPNode{
		inPort: port.NewIn(),
	}
	n.inPort.AddListener(port.ListenFunc(n.forward))
	return n
}

// In returns the input port with the specified name.
func (n *NOPNode) In(name string) *port.InPort {
	switch name {
	case node.PortIn:
		return n.inPort
	default:
		return nil
	}
}

// Out returns nil as NOPNode does not have any output port.
func (n *NOPNode) Out(_ string) *port.OutPort {
	return nil
}

// Close closes all ports associated with the node.
func (n *NOPNode) Close() error {
	n.inPort.Close()
	return nil
}

func (n *NOPNode) forward(proc *process.Process) {
	inReader := n.inPort.Open(proc)

	for range inReader.Read() {
		inReader.Receive(packet.None)
	}
}
