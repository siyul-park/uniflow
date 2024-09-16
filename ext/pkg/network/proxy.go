package network

import (
	"bytes"
	"io"
	"net/http"
	"net/http/httptest"
	"net/http/httputil"
	"net/textproto"
	"net/url"
	"sync"

	"github.com/pkg/errors"
	"github.com/siyul-park/uniflow/ext/pkg/mime"
	"github.com/siyul-park/uniflow/pkg/encoding"
	"github.com/siyul-park/uniflow/pkg/node"
	"github.com/siyul-park/uniflow/pkg/packet"
	"github.com/siyul-park/uniflow/pkg/process"
	"github.com/siyul-park/uniflow/pkg/scheme"
	"github.com/siyul-park/uniflow/pkg/spec"
	"github.com/siyul-park/uniflow/pkg/types"
	"golang.org/x/net/http2"
)

// ProxyNodeSpec defines the specifications for creating a ProxyNode.
type ProxyNodeSpec struct {
	spec.Meta `map:",inline"`
	URLs      []string `map:"urls"`
}

// ProxyNode represents a Node for handling HTTP proxy.
type ProxyNode struct {
	*node.OneToOneNode
	proxy *httputil.ReverseProxy
}

const KindProxy = "proxy"

// NewProxyNodeCodec creates a new codec for ProxyNode.
func NewProxyNodeCodec() scheme.Codec {
	return scheme.CodecWithType(func(spec *ProxyNodeSpec) (node.Node, error) {
		urls := make([]*url.URL, 0, len(spec.URLs))
		for _, u := range spec.URLs {
			parsed, err := url.Parse(u)
			if err != nil {
				return nil, err
			}
			urls = append(urls, parsed)
		}
		if len(urls) == 0 {
			return nil, errors.WithStack(encoding.ErrUnsupportedValue)
		}
		return NewProxyNode(urls), nil
	})
}

// NewProxyNode creates a new ProxyNode instance.
func NewProxyNode(urls []*url.URL) *ProxyNode {
	var index int
	var mu sync.Mutex

	transport := &http.Transport{}
	http2.ConfigureTransport(transport)

	proxy := &httputil.ReverseProxy{
		Transport: transport,
		Rewrite: func(r *httputil.ProxyRequest) {
			mu.Lock()
			defer mu.Unlock()

			index = (index + 1) % len(urls)

			r.SetURL(urls[index])
			r.SetXForwarded()
		},
	}

	n := &ProxyNode{proxy: proxy}
	n.OneToOneNode = node.NewOneToOneNode(n.action)
	return n
}

// action handles the HTTP proxy request and response.
func (n *ProxyNode) action(proc *process.Process, inPck *packet.Packet) (*packet.Packet, *packet.Packet) {
	req := &HTTPPayload{}
	if err := types.Unmarshal(inPck.Payload(), req); err != nil {
		return nil, packet.New(types.NewError(err))
	}

	buf := bytes.NewBuffer(nil)
	if err := mime.Encode(buf, req.Body, textproto.MIMEHeader(req.Header)); err != nil {
		return nil, packet.New(types.NewError(err))
	}

	r := &http.Request{
		Method: req.Method,
		URL: &url.URL{
			Scheme:   req.Scheme,
			Host:     req.Host,
			Path:     req.Path,
			RawQuery: req.Query.Encode(),
		},
		Proto:  req.Protocol,
		Header: req.Header,
		Body:   io.NopCloser(buf),
	}
	w := httptest.NewRecorder()

	n.proxy.ServeHTTP(w, r)

	body, err := mime.Decode(w.Body, textproto.MIMEHeader(w.Header()))
	if err != nil {
		return nil, packet.New(types.NewError(err))
	}

	res := &HTTPPayload{
		Method:   req.Method,
		Scheme:   req.Scheme,
		Host:     req.Host,
		Path:     req.Path,
		Query:    req.Query,
		Protocol: req.Protocol,
		Header:   w.Header(),
		Body:     body,
		Status:   w.Code,
	}

	outPayload, err := types.Marshal(res)
	if err != nil {
		return nil, packet.New(types.NewError(err))
	}
	return packet.New(outPayload), nil
}
