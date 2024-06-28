package net

import (
	"context"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"

	"github.com/siyul-park/uniflow/node"
	"github.com/siyul-park/uniflow/object"
	"github.com/siyul-park/uniflow/packet"
	"github.com/siyul-park/uniflow/port"
	"github.com/siyul-park/uniflow/process"
	"github.com/stretchr/testify/assert"
)

func TestNewHTTPClientNode(t *testing.T) {
	n := NewHTTPClientNode(&url.URL{})
	assert.NotNil(t, n)
	assert.NoError(t, n.Close())
}

func TestHTTPClient_SendAndReceive(t *testing.T) {
	s := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
	}))
	defer s.Close()

	u, _ := url.Parse(s.URL)

	ctx, cancel := context.WithTimeout(context.TODO(), time.Second)
	defer cancel()

	n := NewHTTPClientNode(u)
	defer n.Close()

	n.SetTimeout(time.Second)

	in := port.NewOut()
	in.Link(n.In(node.PortIn))

	proc := process.New()
	defer proc.Exit(nil)

	inWriter := in.Open(proc)

	var inPayload object.Object
	inPck := packet.New(inPayload)

	inWriter.Write(inPck)

	select {
	case outPck := <-inWriter.Receive():
		_, ok := outPck.Payload().(object.Error)
		assert.False(t, ok)
	case <-ctx.Done():
		assert.Fail(t, ctx.Err().Error())
	}
}

func TestHTTPClientNodeCodec_Decode(t *testing.T) {
	s := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
	}))
	defer s.Close()

	codec := NewHTTPClientNodeCodec()

	spec := &HTTPClientNodeSpec{
		URL: "http://localhost:3000/",
	}

	n, err := codec.Decode(spec)
	assert.NoError(t, err)
	assert.NotNil(t, n)
	assert.NoError(t, n.Close())
}

func BenchmarkHTTPClientNode_SendAndReceive(b *testing.B) {
	s := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
	}))
	defer s.Close()

	u, _ := url.Parse(s.URL)

	n := NewHTTPClientNode(u)
	defer n.Close()

	n.SetTimeout(time.Second)

	in := port.NewOut()
	in.Link(n.In(node.PortIn))

	proc := process.New()
	defer proc.Exit(nil)

	inWriter := in.Open(proc)

	inPayload := object.NewMap(
		object.NewString("method"), object.NewString(http.MethodGet),
		object.NewString("url"), object.NewString(s.URL),
	)
	inPck := packet.New(inPayload)

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		inWriter.Write(inPck)
		<-inWriter.Receive()
	}
}
