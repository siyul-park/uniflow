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

func TestNewWebSocketClient(t *testing.T) {
	n := NewWebSocketClientNode()
	assert.NotNil(t, n)
	assert.NoError(t, n.Close())
}

func TestWebSocketClientNode_Port(t *testing.T) {
	n := NewWebSocketClientNode()
	defer n.Close()

	assert.NotNil(t, n.In(node.PortIO))
	assert.NotNil(t, n.In(node.PortIn))
	assert.NotNil(t, n.Out(node.PortOut))
	assert.NotNil(t, n.Out(node.PortErr))
}

func TestWebSocket_URL(t *testing.T) {
	n := NewWebSocketClientNode()
	defer n.Close()

	err := n.SetURL(faker.URL())
	assert.NoError(t, err)
}

func TestWebSocketClientNode_SendAndReceive(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.TODO(), time.Second)
	defer cancel()

	p, err := freeport.GetFreePort()
	assert.NoError(t, err)

	http := NewHTTPServerNode(fmt.Sprintf(":%d", p))
	defer http.Close()

	ws := NewWebSocketUpgradeNode()
	defer ws.Close()

	client := NewWebSocketClientNode()
	defer client.Close()

	http.Out(node.PortOut).Link(ws.In(node.PortIO))
	ws.Out(node.PortOut).Link(ws.In(node.PortIn))

	assert.NoError(t, http.Listen())

	client.SetURL(fmt.Sprintf("ws://localhost:%d", p))

	io := port.NewOut()
	io.Link(client.In(node.PortIO))

	in := port.NewOut()
	in.Link(client.In(node.PortIn))

	out := port.NewIn()
	client.Out(node.PortOut).Link(out)

	proc := process.New()
	defer proc.Close()

	ioWriter := io.Open(proc)
	inWriter := in.Open(proc)
	outReader := out.Open(proc)

	var inPayload primitive.Value
	inPck := packet.New(inPayload)

	ioWriter.Write(inPck)

	select {
	case <-proc.Stack().Done(inPck):
	case <-ctx.Done():
		assert.Fail(t, ctx.Err().Error())
	}

	inPayload, _ = primitive.MarshalBinary(&WebSocketPayload{
		Type: websocket.TextMessage,
		Data: primitive.NewString(faker.UUIDHyphenated()),
	})
	inPck = packet.New(inPayload)

	inWriter.Write(inPck)

	select {
	case outPck := <-outReader.Read():
		proc.Stack().Clear(outPck)
		err, _ := packet.AsError(outPck)
		assert.NoError(t, err)
	case <-ctx.Done():
		assert.Fail(t, ctx.Err().Error())
	}

	inPayload, _ = primitive.MarshalBinary(&WebSocketPayload{
		Type: websocket.CloseMessage,
	})
	inPck = packet.New(inPayload)

	inWriter.Write(inPck)

	select {
	case <-proc.Stack().Done(inPck):
	case <-ctx.Done():
		assert.Fail(t, ctx.Err().Error())
	}
}

func TestWebSocketClientNodeCodec_Decode(t *testing.T) {
	codec := NewWebSocketClientNodeCodec()

	spec := &WebSocketClientNodeSpec{
		URL: "ws://localhost:8080",
	}
	n, err := codec.Decode(spec)
	assert.NoError(t, err)
	assert.NotNil(t, n)
	assert.NoError(t, n.Close())
}

func BenchmarkWebSocketClientNode_SendAndReceive(b *testing.B) {
	p, _ := freeport.GetFreePort()

	http := NewHTTPServerNode(fmt.Sprintf(":%d", p))
	defer http.Close()

	ws := NewWebSocketUpgradeNode()
	defer ws.Close()

	client := NewWebSocketClientNode()
	defer client.Close()

	http.Out(node.PortOut).Link(ws.In(node.PortIO))
	ws.Out(node.PortOut).Link(ws.In(node.PortIn))

	http.Listen()

	client.SetURL(fmt.Sprintf("ws://localhost:%d", p))

	io := port.NewOut()
	io.Link(client.In(node.PortIO))

	in := port.NewOut()
	in.Link(client.In(node.PortIn))

	out := port.NewIn()
	client.Out(node.PortOut).Link(out)

	proc := process.New()
	defer proc.Close()

	ioWriter := io.Open(proc)
	inWriter := in.Open(proc)
	outReader := out.Open(proc)

	var inPayload primitive.Value
	inPck := packet.New(inPayload)

	ioWriter.Write(inPck)

	inPayload, _ = primitive.MarshalBinary(&WebSocketPayload{
		Type: websocket.TextMessage,
		Data: primitive.NewString(faker.UUIDHyphenated()),
	})
	inPck = packet.New(inPayload)

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		inWriter.Write(inPck)
		outPck := <-outReader.Read()
		proc.Stack().Clear(outPck)
	}

	inPayload, _ = primitive.MarshalBinary(&WebSocketPayload{
		Type: websocket.CloseMessage,
	})
	inPck = packet.New(inPayload)

	inWriter.Write(inPck)
}
