package node

import (
	"github.com/siyul-park/uniflow/pkg/node"
	"github.com/siyul-park/uniflow/pkg/packet"
	"github.com/siyul-park/uniflow/pkg/port"
	"github.com/siyul-park/uniflow/pkg/scheme"
	"github.com/siyul-park/uniflow/pkg/spec"
	"github.com/siyul-park/uniflow/pkg/testing"
	"github.com/siyul-park/uniflow/pkg/types"
)

// TestNodeSpec defines the specifications for creating a TestNode.
type TestNodeSpec struct {
	spec.Meta `json:",inline"`
}

// TestNode is a test node implementing node.Node and node.Suite interfaces.
type TestNode struct {
	outPorts [2]*port.OutPort
}

const KindTest = "test"

var (
	_ node.Node     = (*TestNode)(nil)
	_ testing.Suite = (*TestNode)(nil)
)

// NewTestNodeCodec creates and returns a codec for decoding TestNodeSpec.
func NewTestNodeCodec() scheme.Codec {
	return scheme.CodecWithType(func(_ *TestNodeSpec) (node.Node, error) {
		return NewTestNode(), nil
	})
}

// NewTestNode creates and returns a new instance of TestNode.
func NewTestNode() *TestNode {
	return &TestNode{outPorts: [2]*port.OutPort{port.NewOut(), port.NewOut()}}
}

// Run executes the test logic, sending packets through output ports and handling errors.
func (n *TestNode) Run(t *testing.Tester) {
	proc := t.Process()

	writer0 := n.outPorts[0].Open(proc)
	writer1 := n.outPorts[1].Open(proc)

	outPck0 := packet.New(nil)
	backPck0 := packet.Send(writer0, outPck0)
	if backPck0 == packet.None {
		t.Exit(nil)
		return
	}
	if err, ok := backPck0.Payload().(types.Error); ok {
		t.Exit(err.Unwrap())
		return
	}

	outPck1 := packet.New(types.NewSlice(backPck0.Payload(), types.NewInt(-1)))
	backPck1 := packet.Send(writer1, outPck1)
	var err error
	if e, ok := backPck1.Payload().(types.Error); ok {
		err = e.Unwrap()
	}
	t.Exit(err)
}

// In returns nil as this node does not use an input port.
func (n *TestNode) In(_ string) *port.InPort {
	return nil
}

// Out returns the output port specified by the name.
func (n *TestNode) Out(name string) *port.OutPort {
	switch name {
	case node.PortOut:
		return n.outPorts[0]
	default:
		if node.NameOfPort(name) == node.PortOut {
			index, ok := node.IndexOfPort(name)
			if ok && index < len(n.outPorts) {
				return n.outPorts[index]
			}
		}
	}
	return nil
}

// Close closes all output ports of the TestNode.
func (n *TestNode) Close() error {
	for _, outPort := range n.outPorts {
		outPort.Close()
	}
	return nil
}
