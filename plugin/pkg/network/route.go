package network

import (
	"github.com/siyul-park/uniflow/pkg/node"
	"github.com/siyul-park/uniflow/pkg/packet"
	"github.com/siyul-park/uniflow/pkg/process"
)

type RouteNode struct {
	*node.OneToManyNode
}

func NewRouteNode() *RouteNode {
	n := &RouteNode{}
	n.OneToManyNode = node.NewOneToManyNode(n.action)
	return n
}

func (n *RouteNode) action(proc *process.Process, inPck *packet.Packet) ([]*packet.Packet, *packet.Packet) {
	return nil, nil
}
