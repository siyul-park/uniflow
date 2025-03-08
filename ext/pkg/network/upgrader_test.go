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
	"github.com/siyul-park/uniflow/pkg/process"
	"github.com/siyul-park/uniflow/pkg/types"
	"github.com/stretchr/testify/require"
)

func TestUpgradeNodeCodec_Compile(t *testing.T) {
	codec := NewUpgradeNodeCodec()

	spec := &UpgradeNodeSpec{
		Protocol: ProtocolWebsocket,
		Timeout:  time.Second,
		Buffer:   64,
	}
	n, err := codec.Compile(spec)
	require.NoError(t, err)
	require.NotNil(t, n)
	require.NoError(t, n.Close())
}

func TestNewWebSocketUpgradeNode(t *testing.T) {
	n := NewWebSocketUpgradeNode()
	require.NotNil(t, n)
	require.NoError(t, n.Close())
}

func TestWebSocket_Timeout(t *testing.T) {
	n := NewWebSocketUpgradeNode()
	defer n.Close()

	v := time.Second

	n.SetTimeout(v)
	require.Equal(t, v, n.Timeout())
}

func TestWebSocket_ReadBufferSize(t *testing.T) {
	n := NewWebSocketUpgradeNode()
	defer n.Close()

	v := 64

	n.SetReadBufferSize(v)
	require.Equal(t, v, n.ReadBufferSize())
}

func TestWebSocket_WriteBufferSize(t *testing.T) {
	n := NewWebSocketUpgradeNode()
	defer n.Close()

	v := 64

	n.SetWriteBufferSize(v)
	require.Equal(t, v, n.WriteBufferSize())
}

func TestWebSocketUpgradeNode_Port(t *testing.T) {
	n := NewWebSocketUpgradeNode()
	defer n.Close()

	require.NotNil(t, n.In(node.PortIO))
	require.NotNil(t, n.In(node.PortIn))
	require.NotNil(t, n.Out(node.PortOut))
	require.NotNil(t, n.Out(node.PortError))
}

func TestWebSocketUpgradeNode_SendAndReceive(t *testing.T) {
	t.Run("Upgrade", func(t *testing.T) {
		port, err := freeport.GetFreePort()
		require.NoError(t, err)

		http := NewHTTPListenNode(fmt.Sprintf(":%d", port))
		defer http.Close()

		ws := NewWebSocketUpgradeNode()
		defer ws.Close()

		http.Out(node.PortOut).Link(ws.In(node.PortIO))
		ws.Out(node.PortOut).Link(ws.In(node.PortIn))

		require.NoError(t, http.Listen())

		conn, _, err := websocket.DefaultDialer.Dial(fmt.Sprintf("ws://localhost:%d", port), nil)
		require.NoError(t, err)
		defer conn.Close()

		msg := faker.UUIDHyphenated()

		conn.WriteMessage(websocket.TextMessage, []byte(msg))

		typ, p, err := conn.ReadMessage()
		require.NoError(t, err)
		require.Equal(t, []byte(msg), p)
		require.Equal(t, websocket.TextMessage, typ)
	})

	t.Run("IO -> Error -> IO", func(t *testing.T) {
		n := NewWebSocketUpgradeNode()
		defer n.Close()

		io := port.NewOut()
		io.Link(n.In(node.PortIO))

		err := port.NewIn()
		n.Out(node.PortError).Link(err)

		proc := process.New()
		defer proc.Exit(nil)

		ioWriter := io.Open(proc)
		errReader := err.Open(proc)

		inPayload := types.NewString("invalid payload")
		inPck := packet.New(inPayload)

		ioWriter.Write(inPck)

		ctx, cancel := context.WithTimeout(context.TODO(), time.Second)
		defer cancel()

		select {
		case outPck := <-errReader.Read():
			require.NotNil(t, outPck)
			errReader.Receive(outPck)
		case <-ctx.Done():
			require.Fail(t, ctx.Err().Error())
		}

		select {
		case backPck := <-ioWriter.Receive():
			require.NotNil(t, backPck)
		case <-ctx.Done():
			require.Fail(t, ctx.Err().Error())
		}
	})
}

func BenchmarkWebSocketUpgradeNode_SendAndReceive(b *testing.B) {
	port, _ := freeport.GetFreePort()

	http := NewHTTPListenNode(fmt.Sprintf(":%d", port))
	defer http.Close()

	ws := NewWebSocketUpgradeNode()
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
