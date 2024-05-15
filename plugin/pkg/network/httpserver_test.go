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

func TestNewHTTPServerNode(t *testing.T) {
	port, err := freeport.GetFreePort()
	assert.NoError(t, err)

	n := NewHTTPServerNode(fmt.Sprintf(":%d", port))
	assert.NotNil(t, n)
	assert.NoError(t, n.Close())
}

func TestHTTPServerNode_Port(t *testing.T) {
	port, err := freeport.GetFreePort()
	assert.NoError(t, err)

	n := NewHTTPServerNode(fmt.Sprintf(":%d", port))
	defer n.Close()

	assert.NotNil(t, n.In(node.PortIn))
	assert.NotNil(t, n.Out(node.PortOut))
	assert.NotNil(t, n.Out(node.PortErr))
}

func TestHTTPServerNode_ListenAndShutdown(t *testing.T) {
	port, err := freeport.GetFreePort()
	assert.NoError(t, err)

	n := NewHTTPServerNode(fmt.Sprintf(":%d", port))
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

func TestHTTPServerNode_ServeHTTP(t *testing.T) {
	t.Run("Not Linked", func(t *testing.T) {
		n := NewHTTPServerNode("")
		defer n.Close()

		r := httptest.NewRequest(http.MethodGet, "/", nil)
		w := httptest.NewRecorder()

		n.ServeHTTP(w, r)

		assert.Equal(t, http.StatusOK, w.Result().StatusCode)
		assert.Equal(t, "", w.Body.String())
	})

	t.Run("Explicit Response", func(t *testing.T) {
		n := NewHTTPServerNode("")
		defer n.Close()

		in := port.NewOut()
		in.Link(n.In(node.PortIn))

		out := port.NewIn()
		n.Out(node.PortOut).Link(out)

		out.AddHandler(port.HandlerFunc(func(proc *process.Process) {
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
		assert.Equal(t, TextPlainCharsetUTF8, w.Header().Get(HeaderContentType))
		assert.Equal(t, body, w.Body.String())
	})

	t.Run("Implicit Response", func(t *testing.T) {
		n := NewHTTPServerNode("")
		defer n.Close()

		in := port.NewOut()
		in.Link(n.In(node.PortIn))

		out := port.NewIn()
		n.Out(node.PortOut).Link(out)

		out.AddHandler(port.HandlerFunc(func(proc *process.Process) {
			outReader := out.Open(proc)

			for {
				inPck, ok := <-outReader.Read()
				if !ok {
					return
				}

				var req HTTPPayload
				inPayload := inPck.Payload()
				_ = primitive.Unmarshal(inPayload, &req)

				outPck := packet.New(req.Body)

				proc.Stack().Add(inPck, outPck)
				outReader.Receive(outPck)
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
		n := NewHTTPServerNode("")
		defer n.Close()

		in := port.NewOut()
		in.Link(n.In(node.PortIn))

		out := port.NewIn()
		n.Out(node.PortOut).Link(out)

		out.AddHandler(port.HandlerFunc(func(proc *process.Process) {
			outReader := out.Open(proc)

			for {
				inPck, ok := <-outReader.Read()
				if !ok {
					return
				}

				err := errors.New(faker.Sentence())

				errPck := packet.WithError(err, inPck)
				proc.Stack().Add(inPck, errPck)
				outReader.Receive(errPck)
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
		n := NewHTTPServerNode("")
		defer n.Close()

		in := port.NewOut()
		in.Link(n.In(node.PortIn))

		out := port.NewIn()
		n.Out(node.PortOut).Link(out)

		err := port.NewIn()
		n.Out(node.PortErr).Link(err)

		out.AddHandler(port.HandlerFunc(func(proc *process.Process) {
			outReader := out.Open(proc)

			for {
				inPck, ok := <-outReader.Read()
				if !ok {
					return
				}

				err := errors.New(faker.Sentence())

				errPck := packet.WithError(err, inPck)
				proc.Stack().Add(inPck, errPck)
				outReader.Receive(errPck)
			}
		}))
		err.AddHandler(port.HandlerFunc(func(proc *process.Process) {
			errReader := err.Open(proc)

			for {
				inPck, ok := <-errReader.Read()
				if !ok {
					return
				}

				err, _ := packet.AsError(inPck)

				outPck := packet.New(primitive.NewString(err.Error()))
				proc.Stack().Add(inPck, outPck)
				errReader.Receive(outPck)
			}
		}))

		body := faker.Sentence()

		r := httptest.NewRequest(http.MethodGet, "/", strings.NewReader(body))
		w := httptest.NewRecorder()

		n.ServeHTTP(w, r)

		assert.Equal(t, http.StatusOK, w.Result().StatusCode)
		assert.Equal(t, TextPlainCharsetUTF8, w.Header().Get(HeaderContentType))
		assert.NotEmpty(t, w.Body.String())
	})
}

func TestHTTPServerNodeCodec_Decode(t *testing.T) {
	port, err := freeport.GetFreePort()
	assert.NoError(t, err)

	codec := NewHTTPServerNodeCodec()

	spec := &HTTPServerNodeSpec{
		Address: fmt.Sprintf(":%d", port),
	}

	n, err := codec.Decode(spec)
	assert.NoError(t, err)
	assert.NotNil(t, n)
	assert.NoError(t, n.Close())
}

func BenchmarkHTTPServerNode_ServeHTTP(b *testing.B) {
	n := NewHTTPServerNode("")
	defer n.Close()

	in := port.NewOut()
	in.Link(n.In(node.PortIn))

	out := port.NewIn()
	n.Out(node.PortOut).Link(out)

	out.AddHandler(port.HandlerFunc(func(proc *process.Process) {
		outReader := out.Open(proc)

		for {
			inPck, ok := <-outReader.Read()
			if !ok {
				return
			}

			err := errors.New(faker.Sentence())

			errPck := packet.WithError(err, inPck)
			proc.Stack().Add(inPck, errPck)
			outReader.Receive(errPck)
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
