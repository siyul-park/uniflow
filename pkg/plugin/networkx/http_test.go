package networkx

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

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

	_ = n.Close()
}

func TestHTTPNode_Port(t *testing.T) {
	port, err := freeport.GetFreePort()
	assert.NoError(t, err)

	n := NewHTTPNode(HTTPNodeConfig{
		Address: fmt.Sprintf(":%d", port),
	})
	defer func() { _ = n.Close() }()

	p, ok := n.Port(node.PortIO)
	assert.True(t, ok)
	assert.NotNil(t, p)

	p, ok = n.Port(node.PortErr)
	assert.True(t, ok)
	assert.NotNil(t, p)
}

func TestHTTPNode_StartAndClose(t *testing.T) {
	port, err := freeport.GetFreePort()
	assert.NoError(t, err)

	n := NewHTTPNode(HTTPNodeConfig{
		Address: fmt.Sprintf(":%d", port),
	})

	errChan := make(chan error)

	go func() {
		if err := n.Start(); err != nil {
			errChan <- err
		}
	}()

	err = n.WaitForListen(errChan)

	assert.NoError(t, err)
	assert.NoError(t, n.Close())
}

func TestHTTPNode_ServeHTTP(t *testing.T) {
	t.Run("Hello World", func(t *testing.T) {
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

	t.Run("HTTPError", func(t *testing.T) {
		n := NewHTTPNode(HTTPNodeConfig{})
		defer func() { _ = n.Close() }()

		httpErr := NotFound

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

				outPayload, _ := primitive.MarshalText(httpErr)
				outPck := packet.New(outPayload)
				proc.Stack().Link(inPck.ID(), outPck.ID())
				ioStream.Send(outPck)
			}
		}))

		r := httptest.NewRequest(http.MethodGet, "/", nil)
		w := httptest.NewRecorder()

		n.ServeHTTP(w, r)

		assert.Equal(t, httpErr.Status, w.Result().StatusCode)
		assert.Equal(t, TextPlainCharsetUTF8, w.Header().Get(HeaderContentType))
		assert.Equal(t, httpErr.Body.Interface(), w.Body.String())
	})
}
