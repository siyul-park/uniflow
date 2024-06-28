package net

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/go-faker/faker/v4"
	"github.com/gorilla/websocket"
	"github.com/phayes/freeport"
	"github.com/siyul-park/uniflow/node"
	"github.com/siyul-park/uniflow/object"
	"github.com/siyul-park/uniflow/packet"
	"github.com/siyul-park/uniflow/port"
	"github.com/siyul-park/uniflow/process"
	"github.com/stretchr/testify/assert"
)

func TestNewWebSocketUpgraderNode(t *testing.T) {
	n := NewWebSocketUpgraderNode()
	assert.NotNil(t, n)
	assert.NoError(t, n.Close())
}

func TestWebSocket_Timeout(t *testing.T) {
	n := NewWebSocketUpgraderNode()
	defer n.Close()

	v := time.Second

	n.SetTimeout(v)
	assert.Equal(t, v, n.Timeout())
}

func TestWebSocket_ReadBufferSize(t *testing.T) {
	n := NewWebSocketUpgraderNode()
	defer n.Close()

	v := 64

	n.SetReadBufferSize(v)
	assert.Equal(t, v, n.ReadBufferSize())
}

func TestWebSocket_WriteBufferSize(t *testing.T) {
	n := NewWebSocketUpgraderNode()
	defer n.Close()

	v := 64

	n.SetWriteBufferSize(v)
	assert.Equal(t, v, n.WriteBufferSize())
}

func TestWebSocketUpgraderNode_Port(t *testing.T) {
	n := NewWebSocketUpgraderNode()
	defer n.Close()

	assert.NotNil(t, n.In(node.PortIO))
	assert.NotNil(t, n.In(node.PortIn))
	assert.NotNil(t, n.Out(node.PortOut))
	assert.NotNil(t, n.Out(node.PortErr))
}

func TestWebSocketUpgraderNode_SendAndReceive(t *testing.T) {
	t.Run("Upgrade", func(t *testing.T) {
		port, err := freeport.GetFreePort()
		assert.NoError(t, err)

		http := NewHTTPListenerNode(fmt.Sprintf(":%d", port))
		defer http.Close()

		ws := NewWebSocketUpgraderNode()
		defer ws.Close()

		http.Out(node.PortOut).Link(ws.In(node.PortIO))
		ws.Out(node.PortOut).Link(ws.In(node.PortIn))

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
	})

	t.Run("IO -> Error -> IO", func(t *testing.T) {
		n := NewWebSocketUpgraderNode()
		defer n.Close()

		io := port.NewOut()
		io.Link(n.In(node.PortIO))

		err := port.NewIn()
		n.Out(node.PortErr).Link(err)

		proc := process.New()
		defer proc.Exit(nil)

		ioWriter := io.Open(proc)
		errReader := err.Open(proc)

		inPayload := object.NewString("invalid payload")
		inPck := packet.New(inPayload)

		ioWriter.Write(inPck)

		ctx, cancel := context.WithTimeout(context.TODO(), time.Second)
		defer cancel()

		select {
		case outPck := <-errReader.Read():
			assert.NotNil(t, outPck)
			errReader.Receive(outPck)
		case <-ctx.Done():
			assert.Fail(t, "timeout")
		}

		select {
		case backPck := <-ioWriter.Receive():
			assert.NotNil(t, backPck)
		case <-ctx.Done():
			assert.Fail(t, "timeout")
		}
	})
}

func TestUpgraderNodeCodec_Decode(t *testing.T) {
	codec := NewUpgraderNodeCodec()

	spec := &UpgraderNodeSpec{
		Protocol: ProtocolWebsocket,
		Timeout:  time.Second,
		Buffer:   64,
	}
	n, err := codec.Decode(spec)
	assert.NoError(t, err)
	assert.NotNil(t, n)
	assert.NoError(t, n.Close())
}

func BenchmarkWebSocketUpgraderNode_SendAndReceive(b *testing.B) {
	port, _ := freeport.GetFreePort()

	http := NewHTTPListenerNode(fmt.Sprintf(":%d", port))
	defer http.Close()

	ws := NewWebSocketUpgraderNode()
	defer ws.Close()

	http.Out(node.PortOut).Link(ws.In(node.PortIO))
	ws.Out(node.PortOut).Link(ws.In(node.PortIn))

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
