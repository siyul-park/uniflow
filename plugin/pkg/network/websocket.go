package network

import (
	"sync"

	"github.com/siyul-park/uniflow/pkg/node"
	"github.com/siyul-park/uniflow/pkg/port"
	"github.com/siyul-park/uniflow/pkg/process"
)

type WebSocketNode struct {
	ioPort  *port.Port
	inPort  *port.Port
	outPort *port.Port
	errPort *port.Port
	mu      sync.RWMutex
}

var _ node.Node = (*WebSocketNode)(nil)

func NewWebsocketNode() *WebSocketNode {
	n := &WebSocketNode{
		ioPort:  port.New(),
		inPort:  port.New(),
		outPort: port.New(),
		errPort: port.New(),
	}

	n.inPort.AddInitHook(port.InitHookFunc(n.upgrade))

	return n
}

func (n *WebSocketNode) Port(name string) (*port.Port, bool) {
	n.mu.RLock()
	defer n.mu.RUnlock()

	switch name {
	case node.PortIO:
		return n.ioPort, true
	case node.PortIn:
		return n.inPort, true
	case node.PortOut:
		return n.outPort, true
	case node.PortErr:
		return n.errPort, true
	default:
	}

	return nil, false
}

func (n *WebSocketNode) Close() error {
	n.mu.Lock()
	defer n.mu.Unlock()

	n.ioPort.Close()
	n.inPort.Close()
	n.outPort.Close()
	n.errPort.Close()

	return nil
}

func (n *WebSocketNode) upgrade(proc *process.Process) {
	n.mu.RLock()
	defer n.mu.RUnlock()
}
