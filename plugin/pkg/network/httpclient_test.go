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
	"github.com/siyul-park/uniflow/pkg/primitive"
	"github.com/siyul-park/uniflow/pkg/process"
	"github.com/stretchr/testify/assert"
)

func TestNewHTTPClientNode(t *testing.T) {
	n := NewHTTPClientNode()
	assert.NotNil(t, n)
	assert.NoError(t, n.Close())
}

func TestHTTPClient_SendAndReceive(t *testing.T) {
	s := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
	}))
	defer s.Close()

	t.Run("StaticURL", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.TODO(), time.Second)
		defer cancel()

		n := NewHTTPClientNode()
		defer n.Close()

		n.SetTimeout(time.Second)
		n.SetMethod(func(_ any) (string, error) {
			return http.MethodGet, nil
		})
		n.SetURL(func(_ any) (string, error) {
			return s.URL, nil
		})

		in := port.NewOut()
		in.Link(n.In(node.PortIn))

		proc := process.New()
		defer proc.Exit(nil)

		inWriter := in.Open(proc)

		var inPayload primitive.Value
		inPck := packet.New(inPayload)

		inWriter.Write(inPck)

		select {
		case outPck := <-inWriter.Receive():
			err, _ := packet.AsError(outPck)
			assert.NoError(t, err)
		case <-ctx.Done():
			assert.Fail(t, ctx.Err().Error())
		}
	})

	t.Run("DynamicURL", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.TODO(), time.Second)
		defer cancel()

		n := NewHTTPClientNode()
		defer n.Close()

		n.SetTimeout(time.Second)

		in := port.NewOut()
		in.Link(n.In(node.PortIn))

		proc := process.New()
		defer proc.Exit(nil)

		inWriter := in.Open(proc)

		inPayload := primitive.NewMap(
			primitive.NewString("method"), primitive.NewString(http.MethodGet),
			primitive.NewString("url"), primitive.NewString(s.URL),
		)
		inPck := packet.New(inPayload)

		inWriter.Write(inPck)

		select {
		case outPck := <-inWriter.Receive():
			err, _ := packet.AsError(outPck)
			assert.NoError(t, err)
		case <-ctx.Done():
			assert.Fail(t, ctx.Err().Error())
		}
	})

	t.Run("DynamicURLElement", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.TODO(), time.Second)
		defer cancel()

		n := NewHTTPClientNode()
		defer n.Close()

		u, _ := url.Parse(s.URL)

		n.SetTimeout(time.Second)

		in := port.NewOut()
		in.Link(n.In(node.PortIn))

		proc := process.New()
		defer proc.Exit(nil)

		inWriter := in.Open(proc)

		inPayload := primitive.NewMap(
			primitive.NewString("method"), primitive.NewString(http.MethodGet),
			primitive.NewString("scheme"), primitive.NewString(u.Scheme),
			primitive.NewString("host"), primitive.NewString(u.Host),
			primitive.NewString("path"), primitive.NewString(u.Path),
		)
		inPck := packet.New(inPayload)

		inWriter.Write(inPck)

		select {
		case outPck := <-inWriter.Receive():
			err, _ := packet.AsError(outPck)
			assert.NoError(t, err)
		case <-ctx.Done():
			assert.Fail(t, ctx.Err().Error())
		}
	})
}

func TestHTTPClientNodeCodec_Decode(t *testing.T) {
	s := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
	}))
	defer s.Close()

	codec := NewHTTPClientNodeCodec()

	spec := &HTTPClientNodeSpec{
		Method: http.MethodGet,
		URL:    s.URL,
		Query:  `{"foo": "bar"}`,
		Header: `{"foo": "bar"}`,
		Body:   `{"foo": "bar"}`,
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

	n := NewHTTPClientNode()
	defer n.Close()

	n.SetTimeout(time.Second)

	in := port.NewOut()
	in.Link(n.In(node.PortIn))

	proc := process.New()
	defer proc.Exit(nil)

	inWriter := in.Open(proc)

	inPayload := primitive.NewMap(
		primitive.NewString("method"), primitive.NewString(http.MethodGet),
		primitive.NewString("url"), primitive.NewString(s.URL),
	)
	inPck := packet.New(inPayload)

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		inWriter.Write(inPck)
		<-inWriter.Receive()
	}
}
