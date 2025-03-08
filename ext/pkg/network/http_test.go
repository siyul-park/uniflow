package network

import (
	"context"
	"crypto/tls"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"

	"github.com/go-faker/faker/v4"

	"golang.org/x/net/http2"

	"github.com/siyul-park/uniflow/pkg/node"
	"github.com/siyul-park/uniflow/pkg/packet"
	"github.com/siyul-park/uniflow/pkg/port"
	"github.com/siyul-park/uniflow/pkg/process"
	"github.com/siyul-park/uniflow/pkg/types"
	"github.com/stretchr/testify/require"
)

func TestHTTPNodeCodec_Compile(t *testing.T) {
	s := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
	}))
	defer s.Close()

	codec := NewHTTPNodeCodec()

	spec := &HTTPNodeSpec{
		URL: "http://localhost:3000/",
	}

	n, err := codec.Compile(spec)
	require.NoError(t, err)
	require.NotNil(t, n)
	require.NoError(t, n.Close())
}

func TestNewHTTPNode(t *testing.T) {
	n := NewHTTPNode(nil)
	require.NotNil(t, n)
	require.NoError(t, n.Close())
}

func TestHTTPNode_SendAndReceive(t *testing.T) {
	t.Run("HTTP/1.1", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.TODO(), time.Second)
		defer cancel()

		s := httptest.NewServer(http.HandlerFunc(func(_ http.ResponseWriter, req *http.Request) {
			require.Equal(t, "HTTP/1.1", req.Proto)

			body, err := io.ReadAll(req.Body)
			require.NoError(t, err)
			require.NotZero(t, len(body))
		}))
		defer s.Close()

		u, _ := url.Parse(s.URL)

		n := NewHTTPNode(nil)
		defer n.Close()

		n.SetURL(u)
		n.SetTimeout(time.Second)

		in := port.NewOut()
		in.Link(n.In(node.PortIn))

		proc := process.New()
		defer proc.Exit(nil)

		inWriter := in.Open(proc)

		inPayload := types.NewString(faker.UUIDHyphenated())
		inPck := packet.New(inPayload)

		inWriter.Write(inPck)

		select {
		case outPck := <-inWriter.Receive():
			_, ok := outPck.Payload().(types.Error)
			require.False(t, ok)
		case <-ctx.Done():
			require.Fail(t, ctx.Err().Error())
		}
	})

	t.Run("HTTP/2", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.TODO(), time.Second)
		defer cancel()

		s := httptest.NewUnstartedServer(http.HandlerFunc(func(_ http.ResponseWriter, req *http.Request) {
			require.Equal(t, "HTTP/2.0", req.Proto)

			body, err := io.ReadAll(req.Body)
			require.NoError(t, err)
			require.NotZero(t, len(body))
		}))
		_ = http2.ConfigureServer(s.Config, nil)

		s.TLS = &tls.Config{
			NextProtos: []string{"h2"},
		}

		s.StartTLS()
		defer s.Close()

		client := &http.Client{
			Transport: &http2.Transport{
				TLSClientConfig: &tls.Config{
					InsecureSkipVerify: true,
				},
			},
		}
		u, _ := url.Parse(s.URL)

		n := NewHTTPNode(client)
		defer n.Close()

		n.SetURL(u)
		n.SetTimeout(time.Second)

		in := port.NewOut()
		in.Link(n.In(node.PortIn))

		proc := process.New()
		defer proc.Exit(nil)

		inWriter := in.Open(proc)

		inPayload := types.NewString(faker.UUIDHyphenated())
		inPck := packet.New(inPayload)

		inWriter.Write(inPck)

		select {
		case outPck := <-inWriter.Receive():
			_, ok := outPck.Payload().(types.Error)
			require.False(t, ok)
		case <-ctx.Done():
			require.Fail(t, ctx.Err().Error())
		}
	})
}

func BenchmarkHTTPNode_SendAndReceive(b *testing.B) {
	s := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
	}))
	defer s.Close()

	u, _ := url.Parse(s.URL)

	n := NewHTTPNode(nil)
	defer n.Close()

	n.SetURL(u)
	n.SetTimeout(time.Second)

	in := port.NewOut()
	in.Link(n.In(node.PortIn))

	proc := process.New()
	defer proc.Exit(nil)

	inWriter := in.Open(proc)

	inPayload := types.NewMap(
		types.NewString("method"), types.NewString(http.MethodGet),
		types.NewString("url"), types.NewString(s.URL),
	)
	inPck := packet.New(inPayload)

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		inWriter.Write(inPck)
		<-inWriter.Receive()
	}
}
