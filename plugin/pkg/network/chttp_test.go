package network

import (
	"context"
	"github.com/siyul-park/uniflow/pkg/node"
	"github.com/siyul-park/uniflow/pkg/packet"
	"github.com/siyul-park/uniflow/pkg/port"
	"github.com/siyul-park/uniflow/pkg/primitive"
	"github.com/siyul-park/uniflow/pkg/process"
	"github.com/siyul-park/uniflow/plugin/internal/language"
	"github.com/stretchr/testify/assert"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"
)

func TestNewCHTTPNode(t *testing.T) {
	n := NewCHTTPNode()
	assert.NotNil(t, n)
	assert.NoError(t, n.Close())
}

func TestCHTTP_SendAndReceive(t *testing.T) {
	s := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
	}))
	defer s.Close()

	t.Run("Static URL", func(t *testing.T) {
		n := NewCHTTPNode()
		defer n.Close()

		n.SetTimeout(time.Second)
		n.SetLanguage(language.Text)

		err := n.SetMethod(http.MethodGet)
		assert.NoError(t, err)

		err = n.SetURL(s.URL)
		assert.NoError(t, err)

		io := port.NewOut()
		io.Link(n.In(node.PortIO))

		proc := process.New()
		defer proc.Close()

		ioWriter := io.Open(proc)

		var inPayload primitive.Value
		inPck := packet.New(inPayload)

		ioWriter.Write(inPck)

		ctx, cancel := context.WithTimeout(context.TODO(), time.Second)
		defer cancel()

		select {
		case outPck := <-ioWriter.Receive():
			err, _ := packet.AsError(outPck)
			assert.NoError(t, err)
		case <-ctx.Done():
			assert.Fail(t, ctx.Err().Error())
		}
	})

	t.Run("Dynamic URL", func(t *testing.T) {
		n := NewCHTTPNode()
		defer n.Close()

		n.SetTimeout(time.Second)

		io := port.NewOut()
		io.Link(n.In(node.PortIO))

		proc := process.New()
		defer proc.Close()

		ioWriter := io.Open(proc)

		inPayload := primitive.NewMap(
			primitive.NewString("method"), primitive.NewString(http.MethodGet),
			primitive.NewString("url"), primitive.NewString(s.URL),
		)
		inPck := packet.New(inPayload)

		ioWriter.Write(inPck)

		ctx, cancel := context.WithTimeout(context.TODO(), time.Second)
		defer cancel()

		select {
		case outPck := <-ioWriter.Receive():
			err, _ := packet.AsError(outPck)
			assert.NoError(t, err)
		case <-ctx.Done():
			assert.Fail(t, ctx.Err().Error())
		}
	})

	t.Run("Dynamic Divided URL", func(t *testing.T) {
		n := NewCHTTPNode()
		defer n.Close()

		u, _ := url.Parse(s.URL)

		n.SetTimeout(time.Second)

		io := port.NewOut()
		io.Link(n.In(node.PortIO))

		proc := process.New()
		defer proc.Close()

		ioWriter := io.Open(proc)

		inPayload := primitive.NewMap(
			primitive.NewString("method"), primitive.NewString(http.MethodGet),
			primitive.NewString("scheme"), primitive.NewString(u.Scheme),
			primitive.NewString("host"), primitive.NewString(u.Host),
			primitive.NewString("path"), primitive.NewString(u.Path),
		)
		inPck := packet.New(inPayload)

		ioWriter.Write(inPck)

		ctx, cancel := context.WithTimeout(context.TODO(), time.Second)
		defer cancel()

		select {
		case outPck := <-ioWriter.Receive():
			err, _ := packet.AsError(outPck)
			assert.NoError(t, err)
		case <-ctx.Done():
			assert.Fail(t, ctx.Err().Error())
		}
	})

	t.Run("With Query", func(t *testing.T) {
		n := NewCHTTPNode()
		defer n.Close()

		n.SetTimeout(time.Second)

		err := n.SetMethod(http.MethodGet)
		assert.NoError(t, err)

		err = n.SetURL(s.URL)
		assert.NoError(t, err)

		err = n.SetQuery(`{"foo": "bar"}`)
		assert.NoError(t, err)

		io := port.NewOut()
		io.Link(n.In(node.PortIO))

		proc := process.New()
		defer proc.Close()

		ioWriter := io.Open(proc)

		var inPayload primitive.Value
		inPck := packet.New(inPayload)

		ioWriter.Write(inPck)

		ctx, cancel := context.WithTimeout(context.TODO(), time.Second)
		defer cancel()

		select {
		case outPck := <-ioWriter.Receive():
			err, _ := packet.AsError(outPck)
			assert.NoError(t, err)
		case <-ctx.Done():
			assert.Fail(t, ctx.Err().Error())
		}
	})

	t.Run("With Header", func(t *testing.T) {
		n := NewCHTTPNode()
		defer n.Close()

		n.SetTimeout(time.Second)

		err := n.SetMethod(http.MethodGet)
		assert.NoError(t, err)

		err = n.SetURL(s.URL)
		assert.NoError(t, err)

		err = n.SetHeader(`{"foo": "bar"}`)
		assert.NoError(t, err)

		io := port.NewOut()
		io.Link(n.In(node.PortIO))

		proc := process.New()
		defer proc.Close()

		ioWriter := io.Open(proc)

		var inPayload primitive.Value
		inPck := packet.New(inPayload)

		ioWriter.Write(inPck)

		ctx, cancel := context.WithTimeout(context.TODO(), time.Second)
		defer cancel()

		select {
		case outPck := <-ioWriter.Receive():
			err, _ := packet.AsError(outPck)
			assert.NoError(t, err)
		case <-ctx.Done():
			assert.Fail(t, ctx.Err().Error())
		}
	})

	t.Run("With Body", func(t *testing.T) {
		n := NewCHTTPNode()
		defer n.Close()

		n.SetTimeout(time.Second)

		err := n.SetMethod(http.MethodPost)
		assert.NoError(t, err)

		err = n.SetURL(s.URL)
		assert.NoError(t, err)

		err = n.SetBody(`{"foo": "bar"}`)
		assert.NoError(t, err)

		io := port.NewOut()
		io.Link(n.In(node.PortIO))

		proc := process.New()
		defer proc.Close()

		ioWriter := io.Open(proc)

		var inPayload primitive.Value
		inPck := packet.New(inPayload)

		ioWriter.Write(inPck)

		ctx, cancel := context.WithTimeout(context.TODO(), time.Second)
		defer cancel()

		select {
		case outPck := <-ioWriter.Receive():
			err, _ := packet.AsError(outPck)
			assert.NoError(t, err)
		case <-ctx.Done():
			assert.Fail(t, ctx.Err().Error())
		}
	})
}

func TestCHTTPNodeCodec_Decode(t *testing.T) {
	s := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
	}))
	defer s.Close()

	codec := NewCHTTPNodeCodec()

	spec := &CHTTPNodeSpec{
		Method: http.MethodGet,
		URL:    s.URL,
	}

	n, err := codec.Decode(spec)
	assert.NoError(t, err)
	assert.NotNil(t, n)

	assert.NoError(t, n.Close())
}

func BenchmarkCHTTPNode_SendAndReceive(b *testing.B) {
	s := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
	}))
	defer s.Close()

	n := NewCHTTPNode()
	defer n.Close()

	n.SetTimeout(time.Second)

	io := port.NewOut()
	io.Link(n.In(node.PortIO))

	proc := process.New()
	defer proc.Close()

	ioWriter := io.Open(proc)

	inPayload := primitive.NewMap(
		primitive.NewString("method"), primitive.NewString(http.MethodGet),
		primitive.NewString("url"), primitive.NewString(s.URL),
	)
	inPck := packet.New(inPayload)

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		ioWriter.Write(inPck)
		<-ioWriter.Receive()
	}
}
