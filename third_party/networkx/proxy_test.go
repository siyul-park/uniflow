package networkx

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/go-faker/faker/v4"
	"github.com/siyul-park/uniflow/pkg/node"
	"github.com/siyul-park/uniflow/pkg/packet"
	"github.com/siyul-park/uniflow/pkg/port"
	"github.com/siyul-park/uniflow/pkg/primitive"
	"github.com/siyul-park/uniflow/pkg/process"
	"github.com/stretchr/testify/assert"
)

func TestNewProxyNode(t *testing.T) {
	n, err := NewProxyNode(faker.URL())
	assert.NoError(t, err)
	assert.NotNil(t, n)

	assert.NoError(t, n.Close())
}

func TestProxyNode_Send(t *testing.T) {
	called := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called++
	}))
	defer server.Close()

	proxy, _ := NewProxyNode(server.URL)
	defer proxy.Close()

	io := port.New()
	ioPort, _ := proxy.Port(node.PortIO)
	ioPort.Link(io)

	proc := process.New()
	defer proc.Exit(nil)

	ioStream := io.Open(proc)

	inPayload, _ := primitive.MarshalText(HTTPPayload{
		Method: http.MethodGet,
		Path:   "/",
	})
	inPck := packet.New(inPayload)

	ioStream.Send(inPck)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	select {
	case outPck := <-ioStream.Receive():
		assert.Equal(t, 1, called)

		var outPayload HTTPPayload
		err := primitive.Unmarshal(outPck.Payload(), &outPayload)
		assert.NoError(t, err)
	case <-ctx.Done():
		assert.Fail(t, "timeout")
	}
}

func BenchmarkProxyNode_Send(b *testing.B) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
	}))
	defer server.Close()

	proxy, _ := NewProxyNode(server.URL)
	defer proxy.Close()

	io := port.New()
	ioPort, _ := proxy.Port(node.PortIO)
	ioPort.Link(io)

	proc := process.New()
	defer proc.Exit(nil)

	ioStream := io.Open(proc)

	inPayload, _ := primitive.MarshalText(HTTPPayload{
		Method: http.MethodGet,
		Path:   "/",
	})
	inPck := packet.New(inPayload)

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		ioStream.Send(inPck)
		<-ioStream.Receive()
	}
}
