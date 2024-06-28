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
	"github.com/siyul-park/uniflow/pkg/scheme"
	"github.com/siyul-park/uniflow/pkg/spec"
)

// HTTPNode represents a node for making HTTP client requests.
type HTTPNode struct {
	*node.OneToOneNode
	url     *url.URL
	timeout time.Duration
	mu      sync.RWMutex
}

// HTTPNodeSpec holds the specifications for creating an HTTPNode.
type HTTPNodeSpec struct {
	spec.Meta `map:",inline"`
	URL       string        `map:"url"`
	Timeout   time.Duration `map:"timeout,omitempty"`
}

// HTTPPayload is the payload structure for HTTP requests and responses.
type HTTPPayload struct {
	Method string        `map:"method,omitempty"`
	Scheme string        `map:"scheme,omitempty"`
	Host   string        `map:"host,omitempty"`
	Path   string        `map:"path,omitempty"`
	Query  url.Values    `map:"query,omitempty"`
	Proto  string        `map:"proto,omitempty"`
	Header http.Header   `map:"header,omitempty"`
	Body   object.Object `map:"body,omitempty"`
	Status int           `map:"status"`
}

const KindHTTP = "http"

// NewHTTPNode creates a new HTTPNode instance.
func NewHTTPNode(url *url.URL) *HTTPNode {
	n := &HTTPNode{url: url}
	n.OneToOneNode = node.NewOneToOneNode(n.action)

	return n
}

// SetTimeout sets the timeout duration for the HTTP request.
func (n *HTTPNode) SetTimeout(timeout time.Duration) {
	n.mu.Lock()
	defer n.mu.Unlock()

	n.timeout = timeout
}

func (n *HTTPNode) action(proc *process.Process, inPck *packet.Packet) (*packet.Packet, *packet.Packet) {
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

// response processes the HTTP response and returns the payload.
func (n *HTTPNode) response(w *http.Response) (*HTTPPayload, error) {
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

// NewHTTPNodeCodec creates a new codec for HTTPNode.
func NewHTTPNodeCodec() scheme.Codec {
	return scheme.CodecWithType(func(spec *HTTPNodeSpec) (node.Node, error) {
		url, err := url.Parse(spec.URL)
		if err != nil {
			return nil, err
		}

		n := NewHTTPNode(url)
		n.SetTimeout(spec.Timeout)
		return n, nil
	})
}

// NewHTTPPayload creates a new HTTPPayload with the given HTTP status code and optional body.
func NewHTTPPayload(status int, body ...object.Object) *HTTPPayload {
	if len(body) == 0 {
		body = []object.Object{object.NewString(http.StatusText(status))}
	}
	return &HTTPPayload{
		Header: http.Header{},
		Body:   body[0],
		Status: status,
	}
}
