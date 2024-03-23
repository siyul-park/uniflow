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

		n.SetLanguage(language.Text)

		err := n.SetMethod(http.MethodGet)
		assert.NoError(t, err)

		u, _ := url.Parse(s.URL)

		err = n.SetScheme(u.Scheme)
		assert.NoError(t, err)

		err = n.SetHost(u.Host)
		assert.NoError(t, err)

		err = n.SetPath(u.Path)
		assert.NoError(t, err)

		err = n.SetQuery(u.Query())
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

	t.Run("With Query", func(t *testing.T) {
		n := NewCHTTPNode()
		defer n.Close()

		u, _ := url.Parse(s.URL)

		err := n.SetMethod(http.MethodGet)
		assert.NoError(t, err)

		err = n.SetScheme(u.Scheme)
		assert.NoError(t, err)

		err = n.SetHost(u.Host)
		assert.NoError(t, err)

		err = n.SetPath(u.Path)
		assert.NoError(t, err)

		err = n.SetQuery(map[string][]string{
			"foo": {"bar"},
		})
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

		u, _ := url.Parse(s.URL)

		err := n.SetMethod(http.MethodGet)
		assert.NoError(t, err)

		err = n.SetScheme(u.Scheme)
		assert.NoError(t, err)

		err = n.SetHost(u.Host)
		assert.NoError(t, err)

		err = n.SetPath(u.Path)
		assert.NoError(t, err)

		err = n.SetHeader(map[string][]string{
			"foo": {"bar"},
		})
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

		u, _ := url.Parse(s.URL)

		err := n.SetMethod(http.MethodPost)
		assert.NoError(t, err)

		err = n.SetScheme(u.Scheme)
		assert.NoError(t, err)

		err = n.SetHost(u.Host)
		assert.NoError(t, err)

		err = n.SetPath(u.Path)
		assert.NoError(t, err)

		err = n.SetBody(primitive.NewMap(
			primitive.NewString("foo"), primitive.NewSlice(primitive.NewString("bar")),
		))
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
