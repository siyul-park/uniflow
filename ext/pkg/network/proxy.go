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
)

type ProxyNodeSpec struct {
	spec.Meta `map:",inline"`
	URLS      []string `map:"urls"`
}

type ProxyNode struct {
	*node.OneToOneNode
	proxy *httputil.ReverseProxy
}

const KindProxy = "proxy"

func (n *ProxyNode) action(proc *process.Process, inPck *packet.Packet) (*packet.Packet, *packet.Packet) {
	req := HTTPPayload{}
	if err := types.Unmarshal(inPck.Payload(), &req); err != nil {
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
		Header: w.Header(),
		Body:   body,
		Status: w.Code,
	}

	outPayload, err := types.Encoder.Encode(res)
	if err != nil {
		return nil, packet.New(types.NewError(err))
	}

	return packet.New(outPayload), nil
}

func NewProxyNode(urls []*url.URL) *ProxyNode {
	var index int
	var mu sync.Mutex
	proxy := &httputil.ReverseProxy{
		Rewrite: func(pr *httputil.ProxyRequest) {
			mu.Lock()
			defer mu.Unlock()

			index = (index + 1) % len(urls)
			pr.SetURL(urls[index])
			pr.SetXForwarded()
		},
	}

	n := &ProxyNode{proxy: proxy}
	n.OneToOneNode = node.NewOneToOneNode(n.action)
	return n
}

func NewProxyNodeCodec() scheme.Codec {
	return scheme.CodecWithType(func(spec *ProxyNodeSpec) (node.Node, error) {
		urls := make([]*url.URL, 0, len(spec.URLS))
		if len(urls) == 0 {
			return nil, errors.WithStack(encoding.ErrUnsupportedValue)
		}

		for _, u := range spec.URLS {
			parsed, err := url.Parse(u)
			if err != nil {
				return nil, err
			}
			urls = append(urls, parsed)
		}

		return NewProxyNode(urls), nil
	})
}
