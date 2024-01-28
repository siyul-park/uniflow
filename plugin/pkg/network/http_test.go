package network

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/go-faker/faker/v4"
	"github.com/phayes/freeport"
	"github.com/pkg/errors"
	"github.com/siyul-park/uniflow/pkg/node"
	"github.com/siyul-park/uniflow/pkg/packet"
	"github.com/siyul-park/uniflow/pkg/port"
	"github.com/siyul-park/uniflow/pkg/primitive"
	"github.com/siyul-park/uniflow/pkg/process"
	"github.com/stretchr/testify/assert"
)

func TestHTTPNodeCodec_Decode(t *testing.T) {
	port, err := freeport.GetFreePort()
	assert.NoError(t, err)

	codec := NewHTTPNodeCodec()

	spec := &HTTPNodeSpec{
		Address: fmt.Sprintf(":%d", port),
	}

	n, err := codec.Decode(spec)
	assert.NoError(t, err)
	assert.NotNil(t, n)
}

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

	err = n.Listen()
	assert.NoError(t, err)

	_, err = http.Get(fmt.Sprintf("http://127.0.0.1:%d", port))
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

		assert.Equal(t, http.StatusNoContent, w.Result().StatusCode)
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

		body := faker.Sentence()

		r := httptest.NewRequest(http.MethodGet, "/", strings.NewReader(body))
		w := httptest.NewRecorder()

		n.ServeHTTP(w, r)

		assert.Equal(t, http.StatusOK, w.Result().StatusCode)
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

		body := faker.Sentence()

		r := httptest.NewRequest(http.MethodGet, "/", strings.NewReader(body))
		w := httptest.NewRecorder()

		n.ServeHTTP(w, r)

		assert.Equal(t, http.StatusOK, w.Result().StatusCode)
		assert.Equal(t, TextPlainCharsetUTF8, w.Header().Get(HeaderContentType))
		assert.Equal(t, body, w.Body.String())
	})

	t.Run("Error Response", func(t *testing.T) {
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

				err := errors.New(faker.Sentence())

				errPck := packet.WithError(err, inPck)
				proc.Graph().Add(inPck.ID(), errPck.ID())
				ioStream.Send(errPck)
			}
		}))

		r := httptest.NewRequest(http.MethodGet, "/", nil)
		w := httptest.NewRecorder()

		n.ServeHTTP(w, r)

		assert.Equal(t, http.StatusInternalServerError, w.Result().StatusCode)
		assert.Equal(t, TextPlainCharsetUTF8, w.Header().Get(HeaderContentType))
		assert.Equal(t, "Internal Server Error", w.Body.String())
	})

	t.Run("Handel Error Response", func(t *testing.T) {
		n := NewHTTPNode("")
		defer n.Close()

		io := port.New()
		ioPort, _ := n.Port(node.PortIO)
		ioPort.Link(io)

		err := port.New()
		errPort, _ := n.Port(node.PortErr)
		errPort.Link(err)

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

				err := errors.New(req.Body.(primitive.String).String())

				errPck := packet.WithError(err, inPck)
				proc.Graph().Add(inPck.ID(), errPck.ID())
				ioStream.Send(errPck)
			}
		}))
		err.AddInitHook(port.InitHookFunc(func(proc *process.Process) {
			errStream := err.Open(proc)

			for {
				inPck, ok := <-errStream.Receive()
				if !ok {
					return
				}

				err, _ := packet.AsError(inPck)

				outPck := packet.New(primitive.NewString(err.Error()))
				proc.Graph().Add(inPck.ID(), outPck.ID())
				errStream.Send(outPck)
			}
		}))

		body := faker.Sentence()

		r := httptest.NewRequest(http.MethodGet, "/", strings.NewReader(body))
		w := httptest.NewRecorder()

		n.ServeHTTP(w, r)

		assert.Equal(t, http.StatusOK, w.Result().StatusCode)
		assert.Equal(t, TextPlainCharsetUTF8, w.Header().Get(HeaderContentType))
		assert.Equal(t, body, w.Body.String())
	})
}

func BenchmarkHTTPNode_ServeHTTP(b *testing.B) {
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

	body := faker.Sentence()

	b.ResetTimer()

	b.RunParallel(func(p *testing.PB) {
		for p.Next() {
			r := httptest.NewRequest(http.MethodGet, "/", strings.NewReader(body))
			w := httptest.NewRecorder()

			n.ServeHTTP(w, r)
		}
	})
}
