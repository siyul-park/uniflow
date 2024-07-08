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
	"github.com/siyul-park/uniflow/ext/pkg/mime"
	"github.com/siyul-park/uniflow/pkg/node"
	"github.com/siyul-park/uniflow/pkg/packet"
	"github.com/siyul-park/uniflow/pkg/port"
	"github.com/siyul-park/uniflow/pkg/process"
	"github.com/siyul-park/uniflow/pkg/types"
	"github.com/stretchr/testify/assert"
)

func TestNewHTTPListenNode(t *testing.T) {
	port, err := freeport.GetFreePort()
	assert.NoError(t, err)

	n := NewHTTPListenNode(fmt.Sprintf(":%d", port))
	assert.NotNil(t, n)
	assert.NoError(t, n.Close())
}

func TestHTTPListenNode_Port(t *testing.T) {
	port, err := freeport.GetFreePort()
	assert.NoError(t, err)

	n := NewHTTPListenNode(fmt.Sprintf(":%d", port))
	defer n.Close()

	assert.NotNil(t, n.Out(node.PortOut))
	assert.NotNil(t, n.Out(node.PortErr))
}

func TestHTTPListenNode_ListenAndShutdown(t *testing.T) {
	port, err := freeport.GetFreePort()
	assert.NoError(t, err)

	n := NewHTTPListenNode(fmt.Sprintf(":%d", port))
	defer n.Close()

	err = n.Listen()
	assert.NoError(t, err)

	_, err = http.Get(fmt.Sprintf("http://127.0.0.1:%d", port))
	assert.NoError(t, err)

	err = n.Shutdown()
	assert.NoError(t, err)

	err = n.Listen()
	assert.NoError(t, err)

	_, err = http.Get(fmt.Sprintf("http://127.0.0.1:%d", port))
	assert.NoError(t, err)

	err = n.Shutdown()
	assert.NoError(t, err)
}

func TestHTTPListenNode_ServeHTTP(t *testing.T) {
	t.Run("NoResponseGiven", func(t *testing.T) {
		n := NewHTTPListenNode("")
		defer n.Close()

		r := httptest.NewRequest(http.MethodGet, "/", nil)
		w := httptest.NewRecorder()

		n.ServeHTTP(w, r)

		assert.Equal(t, http.StatusOK, w.Result().StatusCode)
		assert.Equal(t, "", w.Body.String())
	})

	t.Run("HTTPPayloadResponse", func(t *testing.T) {
		n := NewHTTPListenNode("")
		defer n.Close()

		out := port.NewIn()
		n.Out(node.PortOut).Link(out)

		out.Accept(port.ListenFunc(func(proc *process.Process) {
			outReader := out.Open(proc)

			for {
				inPck, ok := <-outReader.Read()
				if !ok {
					return
				}

				outReader.Receive(inPck)
			}
		}))

		body := faker.Sentence()

		r := httptest.NewRequest(http.MethodGet, "/", strings.NewReader(body))
		w := httptest.NewRecorder()

		n.ServeHTTP(w, r)

		assert.Equal(t, http.StatusOK, w.Result().StatusCode)
		assert.Equal(t, body, w.Body.String())
	})

	t.Run("BodyResponse", func(t *testing.T) {
		n := NewHTTPListenNode("")
		defer n.Close()

		out := port.NewIn()
		n.Out(node.PortOut).Link(out)

		out.Accept(port.ListenFunc(func(proc *process.Process) {
			outReader := out.Open(proc)

			for {
				inPck, ok := <-outReader.Read()
				if !ok {
					return
				}

				inPayload := inPck.Payload()

				var req *HTTPPayload
				_ = types.Decoder.Decode(inPayload, &req)

				outPck := packet.New(req.Body)
				outReader.Receive(outPck)
			}
		}))

		body := faker.Sentence()

		r := httptest.NewRequest(http.MethodGet, "/", strings.NewReader(body))
		w := httptest.NewRecorder()

		n.ServeHTTP(w, r)

		assert.Equal(t, http.StatusOK, w.Result().StatusCode)
		assert.Equal(t, body, w.Body.String())
	})

	t.Run("ErrorResponse", func(t *testing.T) {
		n := NewHTTPListenNode("")
		defer n.Close()

		out := port.NewIn()
		n.Out(node.PortOut).Link(out)

		out.Accept(port.ListenFunc(func(proc *process.Process) {
			outReader := out.Open(proc)

			for {
				_, ok := <-outReader.Read()
				if !ok {
					return
				}

				err := errors.New(faker.Sentence())

				errPck := packet.New(types.NewError(err))
				outReader.Receive(errPck)
			}
		}))

		r := httptest.NewRequest(http.MethodGet, "/", nil)
		w := httptest.NewRecorder()

		n.ServeHTTP(w, r)

		assert.Equal(t, http.StatusInternalServerError, w.Result().StatusCode)
		assert.Equal(t, "Internal Server Error", w.Body.String())
	})

	t.Run("HandleErrorResponse", func(t *testing.T) {
		n := NewHTTPListenNode("")
		defer n.Close()

		out := port.NewIn()
		n.Out(node.PortOut).Link(out)

		err := port.NewIn()
		n.Out(node.PortErr).Link(err)

		out.Accept(port.ListenFunc(func(proc *process.Process) {
			outReader := out.Open(proc)

			for {
				_, ok := <-outReader.Read()
				if !ok {
					return
				}

				err := errors.New(faker.Sentence())

				errPck := packet.New(types.NewError(err))
				outReader.Receive(errPck)
			}
		}))
		err.Accept(port.ListenFunc(func(proc *process.Process) {
			errReader := err.Open(proc)

			for {
				inPck, ok := <-errReader.Read()
				if !ok {
					return
				}

				err, _ := inPck.Payload().(types.Error)

				outPck := packet.New(types.NewString(err.Error()))
				errReader.Receive(outPck)
			}
		}))

		body := faker.Sentence()

		r := httptest.NewRequest(http.MethodGet, "/", strings.NewReader(body))
		w := httptest.NewRecorder()

		n.ServeHTTP(w, r)

		assert.Equal(t, http.StatusOK, w.Result().StatusCode)
		assert.Equal(t, mime.TextPlainCharsetUTF8, w.Header().Get(mime.HeaderContentType))
		assert.NotEmpty(t, w.Body.String())
	})
}

func TestListenNodeCodec_Decode(t *testing.T) {
	port, err := freeport.GetFreePort()
	assert.NoError(t, err)

	codec := NewListenNodeCodec()

	spec := &ListenNodeSpec{
		Protocol: ProtocolHTTP,
		Port:     port,
	}

	n, err := codec.Compile(spec)
	assert.NoError(t, err)
	assert.NotNil(t, n)
	assert.NoError(t, n.Close())
}

func BenchmarkHTTPListenNode_ServeHTTP(b *testing.B) {
	n := NewHTTPListenNode("")
	defer n.Close()

	in := port.NewOut()
	in.Link(n.In(node.PortIn))

	out := port.NewIn()
	n.Out(node.PortOut).Link(out)

	out.Accept(port.ListenFunc(func(proc *process.Process) {
		outReader := out.Open(proc)

		for {
			inPck, ok := <-outReader.Read()
			if !ok {
				return
			}

			outReader.Receive(inPck)
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
