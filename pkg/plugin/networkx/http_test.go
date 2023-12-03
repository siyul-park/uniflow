package networkx

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/phayes/freeport"
	"github.com/siyul-park/uniflow/pkg/node"
	"github.com/siyul-park/uniflow/pkg/packet"
	"github.com/siyul-park/uniflow/pkg/port"
	"github.com/siyul-park/uniflow/pkg/primitive"
	"github.com/siyul-park/uniflow/pkg/process"
	"github.com/stretchr/testify/assert"
)

func TestNewHTTPNode(t *testing.T) {
	port, err := freeport.GetFreePort()
	assert.NoError(t, err)

	n := NewHTTPNode(HTTPNodeConfig{
		Address: fmt.Sprintf(":%d", port),
	})
	assert.NotNil(t, n)
	assert.NotZero(t, n.ID())

	assert.NoError(t, n.Close())
}

func TestHTTPNode_Port(t *testing.T) {
	port, err := freeport.GetFreePort()
	assert.NoError(t, err)

	n := NewHTTPNode(HTTPNodeConfig{
		Address: fmt.Sprintf(":%d", port),
	})
	defer n.Close()

	p, ok := n.Port(node.PortIO)
	assert.True(t, ok)
	assert.NotNil(t, p)

	p, ok = n.Port(node.PortIn)
	assert.True(t, ok)
	assert.NotNil(t, p)

	p, ok = n.Port(node.PortOut)
	assert.True(t, ok)
	assert.NotNil(t, p)

	p, ok = n.Port(node.PortErr)
	assert.True(t, ok)
	assert.NotNil(t, p)
}

func TestHTTPNode_ServeAndShutdown(t *testing.T) {
	port, err := freeport.GetFreePort()
	assert.NoError(t, err)

	n := NewHTTPNode(HTTPNodeConfig{
		Address: fmt.Sprintf(":%d", port),
	})
	defer n.Close()

	errChan := make(chan error)

	go func() {
		if err := n.Serve(); err != nil {
			errChan <- err
		}
	}()

	err = n.WaitForListen(errChan)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	assert.NoError(t, err)
	assert.NoError(t, n.Shutdown(ctx))
}

func TestHTTPNode_ServeHTTP(t *testing.T) {
	t.Run("IO", func(t *testing.T) {
		n := NewHTTPNode(HTTPNodeConfig{})
		defer func() { _ = n.Close() }()

		io := port.New()
		ioPort, _ := n.Port(node.PortIO)
		ioPort.Link(io)

		io.AddInitHook(port.InitHookFunc(func(proc *process.Process) {
			ioStream := io.Open(proc)

			for {
				inPck, ok := <-ioStream.Receive()
				if !ok {
					return
				}

				outPck := packet.New(primitive.NewMap(
					primitive.NewString("body"), primitive.NewString("Hello World!"),
					primitive.NewString("status"), primitive.NewInt(200),
				))
				proc.Stack().Link(inPck.ID(), outPck.ID())
				ioStream.Send(outPck)
			}
		}))

		r := httptest.NewRequest(http.MethodGet, "/", nil)
		w := httptest.NewRecorder()

		n.ServeHTTP(w, r)

		assert.Equal(t, 200, w.Result().StatusCode)
		assert.Equal(t, TextPlainCharsetUTF8, w.Header().Get(HeaderContentType))
		assert.Equal(t, "Hello World!", w.Body.String())
	})

	t.Run("In/Out", func(t *testing.T) {
		n := NewHTTPNode(HTTPNodeConfig{})
		defer func() { _ = n.Close() }()

		in := port.New()
		inPort, _ := n.Port(node.PortIn)
		inPort.Link(in)

		out := port.New()
		outPort, _ := n.Port(node.PortOut)
		outPort.Link(out)

		out.AddInitHook(port.InitHookFunc(func(proc *process.Process) {
			inStream := in.Open(proc)
			outStream := out.Open(proc)

			for {
				inPck, ok := <-outStream.Receive()
				if !ok {
					return
				}

				outPck := packet.New(primitive.NewMap(
					primitive.NewString("body"), primitive.NewString("Hello World!"),
					primitive.NewString("status"), primitive.NewInt(200),
				))
				proc.Stack().Link(inPck.ID(), outPck.ID())
				inStream.Send(outPck)
			}
		}))

		r := httptest.NewRequest(http.MethodGet, "/", nil)
		w := httptest.NewRecorder()

		n.ServeHTTP(w, r)

		assert.Equal(t, 200, w.Result().StatusCode)
		assert.Equal(t, TextPlainCharsetUTF8, w.Header().Get(HeaderContentType))
		assert.Equal(t, "Hello World!", w.Body.String())
	})
}

func BenchmarkHTTPNode_Send(b *testing.B) {
	b.Run("IO", func(b *testing.B) {
		n := NewHTTPNode(HTTPNodeConfig{})
		defer func() { _ = n.Close() }()

		io := port.New()
		ioPort, _ := n.Port(node.PortIO)
		ioPort.Link(io)

		io.AddInitHook(port.InitHookFunc(func(proc *process.Process) {
			ioStream := io.Open(proc)

			for {
				inPck, ok := <-ioStream.Receive()
				if !ok {
					return
				}

				outPck := packet.New(primitive.NewMap(
					primitive.NewString("body"), primitive.NewString("Hello World!"),
					primitive.NewString("status"), primitive.NewInt(200),
				))
				proc.Stack().Link(inPck.ID(), outPck.ID())
				ioStream.Send(outPck)
			}
		}))

		b.ResetTimer()

		for i := 0; i < b.N; i++ {
			r := httptest.NewRequest(http.MethodGet, "/", nil)
			w := httptest.NewRecorder()

			n.ServeHTTP(w, r)
		}
	})

	b.Run("In/Out", func(b *testing.B) {
		n := NewHTTPNode(HTTPNodeConfig{})
		defer func() { _ = n.Close() }()

		in := port.New()
		inPort, _ := n.Port(node.PortIn)
		inPort.Link(in)

		out := port.New()
		outPort, _ := n.Port(node.PortOut)
		outPort.Link(out)

		out.AddInitHook(port.InitHookFunc(func(proc *process.Process) {
			inStream := in.Open(proc)
			outStream := out.Open(proc)

			for {
				inPck, ok := <-outStream.Receive()
				if !ok {
					return
				}

				outPck := packet.New(primitive.NewMap(
					primitive.NewString("body"), primitive.NewString("Hello World!"),
					primitive.NewString("status"), primitive.NewInt(200),
				))
				proc.Stack().Link(inPck.ID(), outPck.ID())
				inStream.Send(outPck)
			}
		}))

		b.ResetTimer()

		for i := 0; i < b.N; i++ {
			r := httptest.NewRequest(http.MethodGet, "/", nil)
			w := httptest.NewRecorder()

			n.ServeHTTP(w, r)
		}
	})
}
