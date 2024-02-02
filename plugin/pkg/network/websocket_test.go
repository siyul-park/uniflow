package network

import (
	"fmt"
	"testing"
	"time"

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

func TestWebsocket_HandshakeTimeout(t *testing.T) {
	n := NewWebsocketNode()
	defer n.Close()

	v := time.Second

	n.SetHandshakeTimeout(v)
	assert.Equal(t, v, n.HandshakeTimeout())
}

func TestWebsocket_ReadBufferSize(t *testing.T) {
	n := NewWebsocketNode()
	defer n.Close()

	v := 64

	n.SetReadBufferSize(v)
	assert.Equal(t, v, n.ReadBufferSize())
}

func TestWebsocket_WriteBufferSize(t *testing.T) {
	n := NewWebsocketNode()
	defer n.Close()

	v := 64

	n.SetWriteBufferSize(v)
	assert.Equal(t, v, n.WriteBufferSize())
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
	t.Run("Upgrade", func(t *testing.T) {
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
	})
}
