package network

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
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

	n := NewHTTPNode(fmt.Sprintf(":%d", port))
	assert.NotNil(t, n)
	assert.NoError(t, n.Close())
}

func TestHTTPNode_Port(t *testing.T) {
	port, err := freeport.GetFreePort()
	assert.NoError(t, err)

	n := NewHTTPNode(fmt.Sprintf(":%d", port))
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

func TestHTTPNode_ListenAndClose(t *testing.T) {
	port, err := freeport.GetFreePort()
	assert.NoError(t, err)

	n := NewHTTPNode(fmt.Sprintf(":%d", port))

	errChan := make(chan error)

	go func() {
		if err := n.Listen(); err != nil {
			errChan <- err
		}
	}()

	err = n.WaitForListen(errChan)

	assert.NoError(t, err)
	assert.NoError(t, n.Close())
}

func TestHTTPNode_ServeHTTP(t *testing.T) {
	t.Run("Not Linked", func(t *testing.T) {
		n := NewHTTPNode("")
		defer n.Close()

		r := httptest.NewRequest(http.MethodGet, "/", nil)
		w := httptest.NewRecorder()

		n.ServeHTTP(w, r)

		assert.Equal(t, 200, w.Result().StatusCode)
		assert.Equal(t, "", w.Body.String())
	})

	t.Run("Explicit Response", func(t *testing.T) {
		n := NewHTTPNode("")
		defer n.Close()

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

				outPck := packet.New(inPck.Payload())
				proc.Graph().Add(inPck.ID(), outPck.ID())
				ioStream.Send(outPck)
			}
		}))

		body := "Hello, world!"

		r := httptest.NewRequest(http.MethodGet, "/", strings.NewReader(body))
		w := httptest.NewRecorder()

		n.ServeHTTP(w, r)

		assert.Equal(t, 200, w.Result().StatusCode)
		assert.Equal(t, TextPlainCharsetUTF8, w.Header().Get(HeaderContentType))
		assert.Equal(t, body, w.Body.String())
	})

	t.Run("Implicit Response", func(t *testing.T) {
		n := NewHTTPNode("")
		defer n.Close()

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

				var req HTTPPayload
				inPayload := inPck.Payload()
				_ = primitive.Unmarshal(inPayload, &req)

				outPck := packet.New(req.Body)
				proc.Graph().Add(inPck.ID(), outPck.ID())
				ioStream.Send(outPck)
			}
		}))

		body := "Hello, world!"

		r := httptest.NewRequest(http.MethodGet, "/", strings.NewReader(body))
		w := httptest.NewRecorder()

		n.ServeHTTP(w, r)

		assert.Equal(t, 200, w.Result().StatusCode)
		assert.Equal(t, TextPlainCharsetUTF8, w.Header().Get(HeaderContentType))
		assert.Equal(t, body, w.Body.String())
	})
}
