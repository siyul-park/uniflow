package node

import (
	"context"
	"io"
	"net/http"
	"net/textproto"
	"net/url"
	"sync"
	"time"

	"github.com/siyul-park/uniflow/pkg/node"
	"github.com/siyul-park/uniflow/pkg/packet"
	"github.com/siyul-park/uniflow/pkg/process"
	"github.com/siyul-park/uniflow/pkg/scheme"
	"github.com/siyul-park/uniflow/pkg/spec"
	"github.com/siyul-park/uniflow/pkg/types"
	"golang.org/x/net/http2"

	"github.com/siyul-park/uniflow/plugin/net/pkg/mime"
)

// HTTPNodeSpec defines the specifications for creating an HTTPNode.
type HTTPNodeSpec struct {
	spec.Meta `json:",inline"`
	URL       string        `json:"url" validate:"required,url"`
	Timeout   time.Duration `json:"timeout,omitempty"`
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
	Method   string      `json:"method,omitempty"`
	Scheme   string      `json:"scheme,omitempty"`
	Host     string      `json:"host,omitempty"`
	Path     string      `json:"path,omitempty"`
	Query    url.Values  `json:"query,omitempty"`
	Protocol string      `json:"protocol,omitempty"`
	Header   http.Header `json:"header,omitempty"`
	Body     types.Value `json:"body,omitempty"`
	Status   int         `json:"status"`
}

const KindHTTP = "http"

// NewHTTPNodeCodec creates a new codec for HTTPNode.
func NewHTTPNodeCodec() scheme.Codec {
	return scheme.CodecWithType(func(spec *HTTPNodeSpec) (node.Node, error) {
		parse, err := url.Parse(spec.URL)
		if err != nil {
			return nil, err
		}

		transport := &http.Transport{}
		if err := http2.ConfigureTransport(transport); err != nil {
			return nil, err
		}
		client := &http.Client{Transport: transport}

		n := NewHTTPNode(client)
		n.SetURL(parse)
		n.SetTimeout(spec.Timeout)
		return n, nil
	})
}

// NewHTTPNode creates a new HTTPNode instance.
func NewHTTPNode(client *http.Client) *HTTPNode {
	if client == nil {
		client = http.DefaultClient
	}
	n := &HTTPNode{client: client, url: &url.URL{}}
	n.OneToOneNode = node.NewOneToOneNode(n.action)
	return n
}

// SetURL sets the URL for the HTTP request.
func (n *HTTPNode) SetURL(url *url.URL) {
	n.mu.Lock()
	defer n.mu.Unlock()

	n.url = url
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

	ctx := context.Context(proc)
	if n.timeout != 0 {
		var cancel func()
		ctx, cancel = context.WithTimeout(ctx, n.timeout)
		defer cancel()
	}

	var req *HTTPPayload
	if err := types.Unmarshal(inPck.Payload(), &req); err != nil {
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

	header := textproto.MIMEHeader{}
	for k, v := range req.Header {
		header[k] = v
	}

	pr, pw := io.Pipe()

	go mime.Encode(pw, req.Body, header)

	r := &http.Request{
		Method: req.Method,
		URL: &url.URL{
			Scheme:   req.Scheme,
			Host:     req.Host,
			Path:     req.Path,
			RawQuery: req.Query.Encode(),
		},
		Header: req.Header,
		Body:   pr,
	}

	w, err := n.client.Do(r.WithContext(ctx))
	if err != nil {
		return nil, packet.New(types.NewError(err))
	}

	body, err := mime.Decode(w.Body, textproto.MIMEHeader(w.Header))
	if err != nil {
		return nil, packet.New(types.NewError(err))
	}

	if b, ok := body.(types.Buffer); ok {
		proc.AddExitHook(process.ExitFunc(func(err error) {
			_ = b.Close()
		}))
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
