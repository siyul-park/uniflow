package testing

import (
	"github.com/siyul-park/uniflow/pkg/node"
	"github.com/siyul-park/uniflow/pkg/port"
	"github.com/siyul-park/uniflow/pkg/testing"
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
	//TODO implement me
	// 1. open writer0, writer1
	// 2. write outPck0 to writer0, payload is nil
	// 3. receive backPck0 in writer0
	// 4. create outPck1, payload is [backPck0.Payload(), -1]
	// 5. write outPck1 to writer1
	// 6. check write ouPck1 is success, check writer1.Write output
	// 7. if write is fail, check backPck0 payload is error, and exit tester as backPck0 payload
	// 8. if write is success, receive backPck1 in writer1
	// 9. if backPck1 payload is error, exit tester as backPck1 error payload
	// 10 if not, exit tester no error
	panic("implement me")
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
