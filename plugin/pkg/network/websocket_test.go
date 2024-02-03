package network

import (
	"fmt"
	"testing"
	"time"

	"github.com/go-faker/faker/v4"
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
	port, err := freeport.GetFreePort()
	assert.NoError(t, err)

	http := NewHTTPNode(fmt.Sprintf(":%d", port))
	defer http.Close()

	ws := NewWebsocketNode()
	defer ws.Close()

	io1, _ := http.Port(node.PortIO)
	io2, _ := ws.Port(node.PortIO)

	io1.Link(io2)

	in, _ := ws.Port(node.PortIn)
	out, _ := ws.Port(node.PortOut)

	out.Link(in)

	assert.NoError(t, http.Listen())

	conn, _, err := websocket.DefaultDialer.Dial(fmt.Sprintf("ws://localhost:%d", port), nil)
	assert.NoError(t, err)
	defer conn.Close()

	conn.SetWriteDeadline(time.Now().Add(time.Second))
	conn.SetReadDeadline(time.Now().Add(time.Second))

	msg := faker.UUIDHyphenated()

	conn.WriteMessage(websocket.TextMessage, []byte(msg))

	typ, p, err := conn.ReadMessage()
	assert.NoError(t, err)
	assert.Equal(t, []byte(msg), p)
	assert.Equal(t, websocket.TextMessage, typ)
}

func BenchmarkWebsocketNode_SendAndReceive(b *testing.B) {
	port, _ := freeport.GetFreePort()

	http := NewHTTPNode(fmt.Sprintf(":%d", port))
	defer http.Close()

	ws := NewWebsocketNode()
	defer ws.Close()

	io1, _ := http.Port(node.PortIO)
	io2, _ := ws.Port(node.PortIO)

	io1.Link(io2)

	in, _ := ws.Port(node.PortIn)
	out, _ := ws.Port(node.PortOut)

	out.Link(in)

	_ = http.Listen()

	conn, _, _ := websocket.DefaultDialer.Dial(fmt.Sprintf("ws://localhost:%d", port), nil)
	defer conn.Close()

	msg := faker.UUIDHyphenated()

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		conn.WriteMessage(websocket.TextMessage, []byte(msg))
		conn.ReadMessage()
	}
}
