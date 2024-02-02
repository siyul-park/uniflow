package network

import (
	"fmt"
	"testing"

	"github.com/gorilla/websocket"
	"github.com/phayes/freeport"
	"github.com/siyul-park/uniflow/pkg/node"
	"github.com/stretchr/testify/assert"
)

func TestNewWebsocketNode(t *testing.T) {
	n := NewWebsocketNode()
	assert.NotNil(t, n)
	assert.NoError(t, n.Close())
}

func TestWebsocketNode_Port(t *testing.T) {
	n := NewWebsocketNode()
	defer n.Close()

	p, ok := n.Port(node.PortIO)
	assert.True(t, ok)
	assert.NotNil(t, p)

	p, ok = n.Port(node.PortIn)
	assert.True(t, ok)
	assert.NotNil(t, p)

	p, ok = n.Port(node.PortOut)
	assert.True(t, ok)
	assert.NotNil(t, p)

	p, ok = n.Port(node.PortErr)
	assert.True(t, ok)
	assert.NotNil(t, p)
}

func TestWebsocketNode_SendAndReceive(t *testing.T) {
	port, err := freeport.GetFreePort()
	assert.NoError(t, err)

	http := NewHTTPNode(fmt.Sprintf(":%d", port))
	defer http.Close()

	ws := NewWebsocketNode()
	defer ws.Close()

	io1, _ := http.Port(node.PortIO)
	io2, _ := ws.Port(node.PortIO)

	io1.Link(io2)

	assert.NoError(t, http.Listen())

	_, _, err = websocket.DefaultDialer.Dial(fmt.Sprintf("ws://localhost:%d", port), nil)
	assert.NoError(t, err)
}
