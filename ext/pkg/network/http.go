package network

import (
	"bytes"
	"context"
	"net/http"
	"net/textproto"
	"net/url"
	"sync"
	"time"

	"github.com/siyul-park/uniflow/ext/pkg/mime"
	"github.com/siyul-park/uniflow/pkg/node"
	"github.com/siyul-park/uniflow/pkg/packet"
	"github.com/siyul-park/uniflow/pkg/process"
	"github.com/siyul-park/uniflow/pkg/scheme"
	"github.com/siyul-park/uniflow/pkg/spec"
	"github.com/siyul-park/uniflow/pkg/types"
	"golang.org/x/net/http2"
)

// HTTPNodeSpec defines the specifications for creating an HTTPNode.
type HTTPNodeSpec struct {
	spec.Meta `map:",inline"`
	URL       string        `map:"url" validate:"required,url"`
	Timeout   time.Duration `map:"timeout,omitempty"`
}

// HTTPNode represents a node for making HTTP client requests.
type HTTPNode struct {
	*node.OneToOneNode
	client  *http.Client
	url     *url.URL
	timeout time.Duration
	mu      sync.RWMutex
}

// HTTPPayload is the payload structure for HTTP requests and responses.
type HTTPPayload struct {
	Method   string      `map:"method,omitempty"`
	Scheme   string      `map:"scheme,omitempty"`
	Host     string      `map:"host,omitempty"`
	Path     string      `map:"path,omitempty"`
	Query    url.Values  `map:"query,omitempty"`
	Protocol string      `map:"protocol,omitempty"`
	Header   http.Header `map:"header,omitempty"`
	Body     types.Value `map:"body,omitempty"`
	Status   int         `map:"status"`
}

const KindHTTP = "http"

// NewHTTPNodeCodec creates a new codec for HTTPNode.
func NewHTTPNodeCodec() scheme.Codec {
	return scheme.CodecWithType(func(spec *HTTPNodeSpec) (node.Node, error) {
		parse, err := url.Parse(spec.URL)
		if err != nil {
			return nil, err
		}

		n := NewHTTPNode(parse)
		n.SetTimeout(spec.Timeout)
		return n, nil
	})
}

// NewHTTPNode creates a new HTTPNode instance.
func NewHTTPNode(url *url.URL) *HTTPNode {
	transport := &http.Transport{}
	_ = http2.ConfigureTransport(transport)

	client := &http.Client{
		Transport: transport,
	}

	n := &HTTPNode{client: client, url: url}
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

	ctx := proc.Context()
	if n.timeout != 0 {
		var cancel func()
		ctx, cancel = context.WithTimeout(ctx, n.timeout)
		defer cancel()
	}

	req := &HTTPPayload{
		Query:  make(url.Values),
		Header: make(http.Header),
	}
	if err := types.Unmarshal(inPck.Payload(), req); err != nil {
		req.Body = inPck.Payload()
	}

	if req.Method == "" {
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
	for k, v := range n.url.Query() {
		for _, v := range v {
			req.Query.Add(k, v)
		}
	}

	buf := bytes.NewBuffer(nil)
	if err := mime.Encode(buf, req.Body, textproto.MIMEHeader(req.Header)); err != nil {
		return nil, packet.New(types.NewError(err))
	}

	u := &url.URL{
		Scheme:   req.Scheme,
		Host:     req.Host,
		Path:     req.Path,
		RawQuery: req.Query.Encode(),
	}

	r, err := http.NewRequest(req.Method, u.String(), buf)
	if err != nil {
		return nil, packet.New(types.NewError(err))
	}
	r = r.WithContext(ctx)

	w, err := n.client.Do(r)
	if err != nil {
		return nil, packet.New(types.NewError(err))
	}

	proc.AddExitHook(process.ExitFunc(func(err error) {
		_ = w.Body.Close()
	}))

	body, err := mime.Decode(w.Body, textproto.MIMEHeader(w.Header))
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
		Header:   w.Header,
		Body:     body,
		Status:   w.StatusCode,
	}

	outPayload, err := types.Marshal(res)
	if err != nil {
		return nil, packet.New(types.NewError(err))
	}
	return packet.New(outPayload), nil
}

// NewHTTPPayload creates a new HTTPPayload with the given HTTP status code and optional body.
func NewHTTPPayload(status int, bodies ...types.Value) *HTTPPayload {
	var body types.Value = types.NewString(http.StatusText(status))
	if len(bodies) > 0 {
		body = bodies[0]
	}
	return &HTTPPayload{
		Header: http.Header{},
		Body:   body,
		Status: status,
	}
}
