package testing

import (
	"github.com/siyul-park/uniflow/pkg/node"
	"github.com/siyul-park/uniflow/pkg/packet"
	"github.com/siyul-park/uniflow/pkg/port"
	"github.com/siyul-park/uniflow/pkg/process"
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
	// 0. create process to use writer0 and writer1
	proc := process.New()
	defer proc.Exit(nil)

	// 1. open writer0, writer1
	writer0 := n.outPorts[0].Open(proc)
	writer1 := n.outPorts[1].Open(proc)

	// 2. write outPck0 to writer0, payload is nil
	outPck0 := packet.New(nil)
	writer0.Write(outPck0)

	// 3. receive backPck0 in writer0
	backPck0 := <-writer0.Receive()

	// 4. create outPck1, payload is [backPck0.Payload(), -1]
	outPck1 := packet.New(types.NewSlice(backPck0.Payload(), types.NewInt(-1)))

	// 5. write outPck1 to writer1
	count := writer1.Write(outPck1)

	// 6. check write outPck1 is success
	if count == 0 {
		// 7. if write is fail, check backPck0 payload is error, and exit tester as backPck0 payload
		if err, ok := backPck0.Payload().(types.Error); ok {
			t.Exit(err.Unwrap())
			return
		}
	}

	// 8. if write is success, receive backPck1 in writer1
	backPck1 := <-writer1.Receive()

	// 9. if backPck1 payload is error, exit tester as backPck1 error payload
	if err, ok := backPck1.Payload().(types.Error); ok {
		t.Exit(err.Unwrap())
		return
	}

	// 10 if not, exit tester no error
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
