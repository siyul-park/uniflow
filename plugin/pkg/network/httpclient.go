package network

import (
	"bytes"
	"context"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"sync"
	"time"

	"github.com/siyul-park/uniflow/pkg/node"
	"github.com/siyul-park/uniflow/pkg/object"
	"github.com/siyul-park/uniflow/pkg/packet"
	"github.com/siyul-park/uniflow/pkg/process"
	"github.com/siyul-park/uniflow/pkg/spec"
)

// HTTPClientNode represents a node for making HTTP client requests.
type HTTPClientNode struct {
	*node.OneToOneNode
	url     *url.URL
	timeout time.Duration
	mu      sync.RWMutex
}

// HTTPClientNodeSpec holds the specifications for creating an HTTPClientNode.
type HTTPClientNodeSpec struct {
	spec.Meta `map:",inline"`
	URL       string        `map:"url"`
	Timeout   time.Duration `map:"timeout,omitempty"`
}

const KindHTTPClient = "http/client"

// NewHTTPClientNode creates a new HTTPClientNode instance.
func NewHTTPClientNode(url *url.URL) *HTTPClientNode {
	n := &HTTPClientNode{url: url}
	n.OneToOneNode = node.NewOneToOneNode(n.action)

	return n
}

// SetTimeout sets the timeout duration for the HTTP request.
func (n *HTTPClientNode) SetTimeout(timeout time.Duration) {
	n.mu.Lock()
	defer n.mu.Unlock()

	n.timeout = timeout
}

func (n *HTTPClientNode) action(proc *process.Process, inPck *packet.Packet) (*packet.Packet, *packet.Packet) {
	n.mu.RLock()
	defer n.mu.RUnlock()

	ctx := context.Background()
	if n.timeout != 0 {
		var cancel func()
		ctx, cancel = context.WithTimeout(ctx, n.timeout)
		defer cancel()
	}

	req := &HTTPPayload{
		Query:  make(url.Values),
		Header: make(http.Header),
	}
	if err := object.Unmarshal(inPck.Payload(), &req); err != nil {
		req.Body = inPck.Payload()
	}
	if req.Method != "" {
		if req.Body == nil {
			req.Method = http.MethodGet
		} else {
			req.Method = http.MethodPost
		}
	}
	if n.url.Scheme != "" {
		req.Scheme = n.url.Scheme
	}
	if n.url.Host != "" {
		req.Host = n.url.Host
	}
	if n.url.Path != "" {
		req.Path, _ = url.JoinPath(n.url.Path, req.Path)
	}
	if len(n.url.Query()) > 0 {
		for k, v := range n.url.Query() {
			for _, v := range v {
				req.Query.Add(k, v)
			}
		}
	}

	contentType := req.Header.Get(HeaderContentType)
	contentEncoding := req.Header.Get(HeaderContentEncoding)

	b, err := MarshalMIME(req.Body, &contentType)
	if err != nil {
		return nil, packet.New(object.NewError(err))
	}
	b, err = Compress(b, contentEncoding)
	if err != nil {
		return nil, packet.New(object.NewError(err))
	}

	req.Header.Set(HeaderContentLength, strconv.Itoa(len(b)))
	if contentType != "" {
		req.Header.Set(HeaderContentType, contentType)
	}

	u := &url.URL{
		Scheme:   req.Scheme,
		Host:     req.Host,
		Path:     req.Path,
		RawQuery: req.Query.Encode(),
	}

	r, err := http.NewRequest(req.Method, u.String(), bytes.NewReader(b))
	if err != nil {
		return nil, packet.New(object.NewError(err))
	}
	r = r.WithContext(ctx)

	client := &http.Client{}

	w, err := client.Do(r)
	if err != nil {
		return nil, packet.New(object.NewError(err))
	}
	defer w.Body.Close()

	res, err := n.response(w)
	if err != nil {
		return nil, packet.New(object.NewError(err))
	}

	outPayload, err := object.MarshalText(res)
	if err != nil {
		return nil, packet.New(object.NewError(err))
	}
	return packet.New(outPayload), nil
}

func (n *HTTPClientNode) response(w *http.Response) (*HTTPPayload, error) {
	contentType := w.Header.Get(HeaderContentType)
	contentEncoding := w.Header.Get(HeaderContentEncoding)

	if b, err := io.ReadAll(w.Body); err != nil {
		return nil, err
	} else if b, err := Decompress(b, contentEncoding); err != nil {
		return nil, err
	} else if b, err := UnmarshalMIME(b, &contentType); err != nil {
		return nil, err
	} else {
		w.Header.Set(HeaderContentType, contentType)

		return &HTTPPayload{
			Header: w.Header,
			Body:   b,
		}, nil
	}
}

func NewHTTPClientNodeCodec() spec.Codec {
	return spec.CodecWithType(func(spec *HTTPClientNodeSpec) (node.Node, error) {
		url, err := url.Parse(spec.URL)
		if err != nil {
			return nil, err
		}

		n := NewHTTPClientNode(url)
		n.SetTimeout(spec.Timeout)
		return n, nil
	})
}
