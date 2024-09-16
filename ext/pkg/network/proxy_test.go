package network

import (
	"context"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"

	"github.com/siyul-park/uniflow/pkg/node"
	"github.com/siyul-park/uniflow/pkg/packet"
	"github.com/siyul-park/uniflow/pkg/port"
	"github.com/siyul-park/uniflow/pkg/process"
	"github.com/siyul-park/uniflow/pkg/types"
	"github.com/stretchr/testify/assert"
)

func TestProxyNodeCodec_Decode(t *testing.T) {
	codec := NewProxyNodeCodec()

	spec := &ProxyNodeSpec{
		URLs: []string{"http://localhost"},
	}

	n, err := codec.Compile(spec)
	assert.NoError(t, err)
	assert.NotNil(t, n)
	assert.NoError(t, n.Close())
}

func TestNewProxyNode(t *testing.T) {
	u, _ := url.Parse("http://localhost")
	n := NewProxyNode([]*url.URL{u})
	assert.NotNil(t, n)
	assert.NoError(t, n.Close())
}

func TestProxyNode_SendAndReceive(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.TODO(), time.Second)
	defer cancel()

	s1 := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		writer.WriteHeader(http.StatusOK)
		writer.Write([]byte("Backend 1"))
	}))
	defer s1.Close()

	s2 := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		writer.WriteHeader(http.StatusOK)
		writer.Write([]byte("Backend 2"))
	}))
	defer s2.Close()

	u1, _ := url.Parse(s1.URL)
	u2, _ := url.Parse(s2.URL)

	n := NewProxyNode([]*url.URL{u1, u2})
	defer n.Close()

	in := port.NewOut()
	in.Link(n.In(node.PortIn))

	proc := process.New()
	defer proc.Exit(nil)

	inWriter := in.Open(proc)

	inPayload := types.NewMap(
		types.NewString("method"), types.NewString(http.MethodGet),
		types.NewString("scheme"), types.NewString("http"),
		types.NewString("host"), types.NewString("test"),
		types.NewString("path"), types.NewString("/"),
		types.NewString("protocol"), types.NewString("HTTP/1.1"),
		types.NewString("status"), types.NewInt(0),
	)
	inPck := packet.New(inPayload)

	inWriter.Write(inPck)

	select {
	case outPck := <-inWriter.Receive():
		payload := &HTTPPayload{}
		err := types.Unmarshal(outPck.Payload(), payload)
		assert.NoError(t, err)
		assert.Contains(t, payload.Body.Interface(), "Backend")
	case <-ctx.Done():
		assert.Fail(t, ctx.Err().Error())
	}
}

func BenchmarkProxyNode_SendAndReceive(b *testing.B) {
	s := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		writer.WriteHeader(http.StatusOK)
		writer.Write([]byte("OK"))
	}))
	defer s.Close()

	u, _ := url.Parse(s.URL)

	n := NewProxyNode([]*url.URL{u})
	defer n.Close()

	in := port.NewOut()
	in.Link(n.In(node.PortIn))

	proc := process.New()
	defer proc.Exit(nil)

	inWriter := in.Open(proc)

	inPayload := types.NewMap(
		types.NewString("method"), types.NewString(http.MethodGet),
		types.NewString("scheme"), types.NewString("http"),
		types.NewString("host"), types.NewString("test"),
		types.NewString("path"), types.NewString("/"),
		types.NewString("protocol"), types.NewString("HTTP/1.1"),
		types.NewString("status"), types.NewInt(0),
	)
	inPck := packet.New(inPayload)

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		inWriter.Write(inPck)
		<-inWriter.Receive()
	}
}
