package testing

import (
	"errors"

	"github.com/siyul-park/uniflow/pkg/node"
	"github.com/siyul-park/uniflow/pkg/packet"
	"github.com/siyul-park/uniflow/pkg/port"
	"github.com/siyul-park/uniflow/pkg/testing"
	"github.com/siyul-park/uniflow/pkg/types"
)

type TestNode struct {
	outPorts [2]*port.OutPort
}

var _ node.Node = (*TestNode)(nil)
var _ testing.Suite = (*TestNode)(nil)

func NewTestNode() *TestNode {
	n := &TestNode{
		outPorts: [2]*port.OutPort{port.NewOut(), port.NewOut()},
	}
	return n
}

func (n *TestNode) Run(t *testing.Tester) {
	proc := t.Process()

	writer0 := n.outPorts[0].Open(proc)
	writer1 := n.outPorts[1].Open(proc)

	outPck0 := packet.New(nil)
	backPck0 := packet.Send(writer0, outPck0)
	if backPck0 == packet.None {
		t.Exit(errors.New("no response from first port"))
		return
	}
	if err, ok := backPck0.Payload().(types.Error); ok {
		t.Exit(err.Unwrap())
		return
	}

	outPck1 := packet.New(types.NewSlice(backPck0.Payload(), types.NewInt32(-1)))
	backPck1 := packet.Send(writer1, outPck1)
	if backPck1 == packet.None {
		t.Exit(errors.New("no response from second port"))
		return
	}
	if err, ok := backPck1.Payload().(types.Error); ok {
		t.Exit(err.Unwrap())
		return
	}

	t.Exit(nil)
}

func (n *TestNode) In(_ string) *port.InPort {
	return nil
}

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

func (n *TestNode) Close() error {
	for _, outPort := range n.outPorts {
		outPort.Close()
	}
	return nil
}
