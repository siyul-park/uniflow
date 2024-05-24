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
	"github.com/siyul-park/uniflow/plugin/internal/language"
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
		n.SetLanguage(language.Text)

		err := n.SetMethod(http.MethodGet)
		assert.NoError(t, err)

		err = n.SetURL(s.URL)
		assert.NoError(t, err)

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

	t.Run("StaticQuery", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.TODO(), time.Second)
		defer cancel()

		n := NewHTTPClientNode()
		defer n.Close()

		n.SetTimeout(time.Second)

		err := n.SetMethod(http.MethodGet)
		assert.NoError(t, err)

		err = n.SetURL(s.URL)
		assert.NoError(t, err)

		err = n.SetQuery(`{"foo": "bar"}`)
		assert.NoError(t, err)

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

	t.Run("StaticHeader", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.TODO(), time.Second)
		defer cancel()

		n := NewHTTPClientNode()
		defer n.Close()

		n.SetTimeout(time.Second)

		err := n.SetMethod(http.MethodGet)
		assert.NoError(t, err)

		err = n.SetURL(s.URL)
		assert.NoError(t, err)

		err = n.SetHeader(`{"foo": "bar"}`)
		assert.NoError(t, err)

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

	t.Run("StaticBody", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.TODO(), time.Second)
		defer cancel()

		n := NewHTTPClientNode()
		defer n.Close()

		n.SetTimeout(time.Second)

		err := n.SetMethod(http.MethodPost)
		assert.NoError(t, err)

		err = n.SetURL(s.URL)
		assert.NoError(t, err)

		err = n.SetBody(`{"foo": "bar"}`)
		assert.NoError(t, err)

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
}

func TestHTTPClientNodeCodec_Decode(t *testing.T) {
	s := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
	}))
	defer s.Close()

	codec := NewHTTPClientNodeCodec()

	spec := &HTTPClientNodeSpec{
		Method: http.MethodGet,
		URL:    s.URL,
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
