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

func TestNewHTTPNode(t *testing.T) {
	n := NewHTTPNode(&url.URL{})
	assert.NotNil(t, n)
	assert.NoError(t, n.Close())
}

func TestHTTP_SendAndReceive(t *testing.T) {
	s := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
	}))
	defer s.Close()

	u, _ := url.Parse(s.URL)

	ctx, cancel := context.WithTimeout(context.TODO(), time.Second)
	defer cancel()

	n := NewHTTPNode(u)
	defer n.Close()

	n.SetTimeout(time.Second)

	in := port.NewOut()
	in.Link(n.In(node.PortIn))

	proc := process.New()
	defer proc.Exit(nil)

	inWriter := in.Open(proc)

	var inPayload types.Value
	inPck := packet.New(inPayload)

	inWriter.Write(inPck)

	select {
	case outPck := <-inWriter.Receive():
		_, ok := outPck.Payload().(types.Error)
		assert.False(t, ok)
	case <-ctx.Done():
		assert.Fail(t, ctx.Err().Error())
	}
}

func TestHTTPNodeCodec_Decode(t *testing.T) {
	s := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
	}))
	defer s.Close()

	codec := NewHTTPNodeCodec()

	spec := &HTTPNodeSpec{
		URL: "http://localhost:3000/",
	}

	n, err := codec.Compile(spec)
	assert.NoError(t, err)
	assert.NotNil(t, n)
	assert.NoError(t, n.Close())
}

func BenchmarkHTTPNode_SendAndReceive(b *testing.B) {
	s := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
	}))
	defer s.Close()

	u, _ := url.Parse(s.URL)

	n := NewHTTPNode(u)
	defer n.Close()

	n.SetTimeout(time.Second)

	in := port.NewOut()
	in.Link(n.In(node.PortIn))

	proc := process.New()
	defer proc.Exit(nil)

	inWriter := in.Open(proc)

	inPayload := types.NewMap(
		types.NewString("method"), types.NewString(http.MethodGet),
		types.NewString("url"), types.NewString(s.URL),
	)
	inPck := packet.New(inPayload)

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		inWriter.Write(inPck)
		<-inWriter.Receive()
	}
}
