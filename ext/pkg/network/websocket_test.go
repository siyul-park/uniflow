package network

import (
	"context"
	"fmt"
	"net/url"
	"testing"
	"time"

	"github.com/go-faker/faker/v4"
	"github.com/gorilla/websocket"
	"github.com/phayes/freeport"
	"github.com/siyul-park/uniflow/pkg/node"
	"github.com/siyul-park/uniflow/pkg/packet"
	"github.com/siyul-park/uniflow/pkg/port"
	"github.com/siyul-park/uniflow/pkg/process"
	"github.com/siyul-park/uniflow/pkg/types"
	"github.com/stretchr/testify/require"
)

func TestWebSocketNodeCodec_Compile(t *testing.T) {
	codec := NewWebSocketNodeCodec()

	spec := &WebSocketNodeSpec{
		URL: "ws://localhost:8080/",
	}
	n, err := codec.Compile(spec)
	require.NoError(t, err)
	require.NotNil(t, n)
	require.NoError(t, n.Close())
}

func TestNewWebSocketClient(t *testing.T) {
	n := NewWebSocketNode(&url.URL{})
	require.NotNil(t, n)
	require.NoError(t, n.Close())
}

func TestWebSocketNode_Port(t *testing.T) {
	n := NewWebSocketNode(&url.URL{})
	defer n.Close()

	require.NotNil(t, n.In(node.PortIO))
	require.NotNil(t, n.In(node.PortIn))
	require.NotNil(t, n.Out(node.PortOut))
	require.NotNil(t, n.Out(node.PortError))
}

func TestWebSocketNode_SendAndReceive(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.TODO(), time.Second)
	defer cancel()

	p, err := freeport.GetFreePort()
	require.NoError(t, err)

	u, _ := url.Parse(fmt.Sprintf("ws://localhost:%d", p))

	http := NewHTTPListenNode(fmt.Sprintf(":%d", p))
	defer http.Close()

	ws := NewWebSocketUpgradeNode()
	defer ws.Close()

	client := NewWebSocketNode(u)
	defer client.Close()

	http.Out(node.PortOut).Link(ws.In(node.PortIO))
	ws.Out(node.PortOut).Link(ws.In(node.PortIn))

	require.NoError(t, http.Listen())

	io := port.NewOut()
	io.Link(client.In(node.PortIO))

	in := port.NewOut()
	in.Link(client.In(node.PortIn))

	out := port.NewIn()
	client.Out(node.PortOut).Link(out)

	proc := process.New()
	defer proc.Exit(nil)

	ioWriter := io.Open(proc)
	inWriter := in.Open(proc)

	done := make(chan struct{})
	out.AddListener(port.ListenFunc(func(proc *process.Process) {
		outReader := out.Open(proc)

		for {
			_, ok := <-outReader.Read()
			if !ok {
				return
			}

			outReader.Receive(packet.None)

			select {
			case <-done:
			default:
				close(done)
			}
		}
	}))

	var inPayload types.Value
	inPck := packet.New(inPayload)

	ioWriter.Write(inPck)

	select {
	case <-ioWriter.Receive():
	case <-ctx.Done():
		require.Fail(t, ctx.Err().Error())
	}

	inPayload, _ = types.Marshal(&WebSocketPayload{
		Type: websocket.TextMessage,
		Data: types.NewString(faker.UUIDHyphenated()),
	})
	inPck = packet.New(inPayload)

	inWriter.Write(inPck)

	select {
	case <-done:
	case <-ctx.Done():
		require.Fail(t, ctx.Err().Error())
	}

	inPayload, _ = types.Marshal(&WebSocketPayload{
		Type: websocket.CloseMessage,
	})
	inPck = packet.New(inPayload)

	inWriter.Write(inPck)

	select {
	case <-inWriter.Receive():
	case <-ctx.Done():
		require.Fail(t, ctx.Err().Error())
	}
}

func BenchmarkWebSocketNode_SendAndReceive(b *testing.B) {
	p, _ := freeport.GetFreePort()

	u, _ := url.Parse(fmt.Sprintf("ws://localhost:%d", p))

	http := NewHTTPListenNode(fmt.Sprintf(":%d", p))
	defer http.Close()

	ws := NewWebSocketUpgradeNode()
	defer ws.Close()

	client := NewWebSocketNode(u)
	defer client.Close()

	http.Out(node.PortOut).Link(ws.In(node.PortIO))
	ws.Out(node.PortOut).Link(ws.In(node.PortIn))

	http.Listen()

	io := port.NewOut()
	io.Link(client.In(node.PortIO))

	in := port.NewOut()
	in.Link(client.In(node.PortIn))

	out := port.NewIn()
	client.Out(node.PortOut).Link(out)

	proc := process.New()
	defer proc.Exit(nil)

	ioWriter := io.Open(proc)
	inWriter := in.Open(proc)

	out.AddListener(port.ListenFunc(func(proc *process.Process) {
		outReader := out.Open(proc)

		for {
			_, ok := <-outReader.Read()
			if !ok {
				return
			}

			outReader.Receive(packet.None)
		}
	}))

	var inPayload types.Value
	inPck := packet.New(inPayload)

	ioWriter.Write(inPck)

	inPayload, _ = types.Marshal(&WebSocketPayload{
		Type: websocket.TextMessage,
		Data: types.NewString(faker.UUIDHyphenated()),
	})
	inPck = packet.New(inPayload)

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		inWriter.Write(inPck)
		<-inWriter.Receive()
	}

	inPayload, _ = types.Marshal(&WebSocketPayload{
		Type: websocket.CloseMessage,
	})
	inPck = packet.New(inPayload)

	inWriter.Write(inPck)
}
