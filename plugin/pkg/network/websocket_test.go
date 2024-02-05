package network

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/go-faker/faker/v4"
	"github.com/gorilla/websocket"
	"github.com/phayes/freeport"
	"github.com/siyul-park/uniflow/pkg/node"
	"github.com/siyul-park/uniflow/pkg/packet"
	"github.com/siyul-park/uniflow/pkg/port"
	"github.com/siyul-park/uniflow/pkg/primitive"
	"github.com/siyul-park/uniflow/pkg/process"
	"github.com/stretchr/testify/assert"
)

func TestWebsocketNodeCodec_Decode(t *testing.T) {
	codec := NewWebsocketNodeCodec()

	spec := &WebsocketNodeSpec{
		Timeout: time.Second,
		Read:    64,
		Write:   64,
	}
	n, err := codec.Decode(spec)
	assert.NoError(t, err)
	assert.NotNil(t, n)
}

func TestNewWebsocketNode(t *testing.T) {
	n := NewWebsocketNode()
	assert.NotNil(t, n)
	assert.NoError(t, n.Close())
}

func TestWebsocket_Timeout(t *testing.T) {
	n := NewWebsocketNode()
	defer n.Close()

	v := time.Second

	n.SetTimeout(v)
	assert.Equal(t, v, n.Timeout())
}

func TestWebsocket_ReadBuffer(t *testing.T) {
	n := NewWebsocketNode()
	defer n.Close()

	v := 64

	n.SetReadBuffer(v)
	assert.Equal(t, v, n.ReadBuffer())
}

func TestWebsocket_WriteBuffer(t *testing.T) {
	n := NewWebsocketNode()
	defer n.Close()

	v := 64

	n.SetWriteBuffer(v)
	assert.Equal(t, v, n.WriteBuffer())
}

func TestWebsocketNode_Port(t *testing.T) {
	n := NewWebsocketNode()
	defer n.Close()

	p := n.Port(node.PortIO)
	assert.NotNil(t, p)

	p = n.Port(node.PortIn)
	assert.NotNil(t, p)

	p = n.Port(node.PortOut)
	assert.NotNil(t, p)

	p = n.Port(node.PortErr)
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

		io1 := http.Port(node.PortIO)
		io2 := ws.Port(node.PortIO)

		io1.Link(io2)

		in := ws.Port(node.PortIn)
		out := ws.Port(node.PortOut)

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
	})

	t.Run("With Error", func(t *testing.T) {
		n := NewWebsocketNode()
		defer n.Close()

		io := port.New()
		ioPort := n.Port(node.PortIO)
		ioPort.Link(io)

		err := port.New()
		errPort := n.Port(node.PortErr)
		errPort.Link(err)

		proc := process.New()
		defer proc.Exit(nil)
		defer proc.Stack().Close()

		ioStream := io.Open(proc)
		errStream := err.Open(proc)

		inPayload := primitive.NewString("invalid payload")
		inPck := packet.New(inPayload)

		ioStream.Send(inPck)

		ctx, cancel := context.WithTimeout(context.TODO(), time.Second)
		defer cancel()

		select {
		case outPck := <-errStream.Receive():
			assert.NotNil(t, outPck)
			errStream.Send(outPck)
		case <-ctx.Done():
			assert.Fail(t, "timeout")
		}

		select {
		case backPck := <-ioStream.Receive():
			assert.NotNil(t, backPck)
		case <-ctx.Done():
			assert.Fail(t, "timeout")
		}
	})
}

func BenchmarkWebsocketNode_SendAndReceive(b *testing.B) {
	port, _ := freeport.GetFreePort()

	http := NewHTTPNode(fmt.Sprintf(":%d", port))
	defer http.Close()

	ws := NewWebsocketNode()
	defer ws.Close()

	io1 := http.Port(node.PortIO)
	io2 := ws.Port(node.PortIO)

	io1.Link(io2)

	in := ws.Port(node.PortIn)
	out := ws.Port(node.PortOut)

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
