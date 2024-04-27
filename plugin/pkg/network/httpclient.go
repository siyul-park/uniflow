package network

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"sync"
	"time"

	"github.com/siyul-park/uniflow/pkg/node"
	"github.com/siyul-park/uniflow/pkg/packet"
	"github.com/siyul-park/uniflow/pkg/primitive"
	"github.com/siyul-park/uniflow/pkg/process"
	"github.com/siyul-park/uniflow/pkg/scheme"
	"github.com/siyul-park/uniflow/plugin/internal/language"
)

// HTTPClientNode represents a node for making HTTP client requests.
type HTTPClientNode struct {
	*node.OneToOneNode
	lang    string
	method  func(primitive.Value) (string, error)
	url     func(primitive.Value) (string, error)
	query   func(primitive.Value) (url.Values, error)
	header  func(primitive.Value) (http.Header, error)
	body    func(primitive.Value) (primitive.Value, error)
	timeout time.Duration
	mu      sync.RWMutex
}

// HTTPClientNodeSpec holds the specifications for creating an HTTPClientNode.
type HTTPClientNodeSpec struct {
	scheme.SpecMeta `map:",inline"`
	Lang            string `map:"lang,omitempty"`
	Method          string `map:"method,omitempty"`
	URL             string `map:"url,omitempty"`
	Query           string `map:"query,omitempty"`
	Header          string `map:"header,omitempty"`
	Body            string `map:"body,omitempty"`
}

const KindHTTPClient = "http/client"

// NewHTTPClientNode creates a new HTTPClientNode instance.
func NewHTTPClientNode() *HTTPClientNode {
	n := &HTTPClientNode{}
	n.OneToOneNode = node.NewOneToOneNode(n.action)

	_ = n.SetMethod("")
	_ = n.SetURL("")
	_ = n.SetQuery("")
	_ = n.SetHeader("")
	_ = n.SetBody("")

	return n
}

// SetLanguage sets the language for transformation.
func (n *HTTPClientNode) SetLanguage(lang string) {
	n.mu.Lock()
	defer n.mu.Unlock()

	n.lang = lang
}

// SetMethod sets the HTTP request method.
func (n *HTTPClientNode) SetMethod(method string) error {
	n.mu.Lock()
	defer n.mu.Unlock()

	if method == "" {
		n.method = func(value primitive.Value) (string, error) {
			if method, ok := primitive.Pick[string](value, "method"); ok {
				return method, nil
			}
			return http.MethodGet, nil
		}
		return nil
	}

	transform, err := language.CompileTransformWithPrimitive(method, n.lang)
	if err != nil {
		return err
	}

	n.method = func(value primitive.Value) (string, error) {
		if v, err := transform(value); err != nil {
			return "", err
		} else {
			return fmt.Sprintf("%v", v.Interface()), nil
		}
	}
	return nil
}

// SetURL sets the target URL for the HTTP request.
func (n *HTTPClientNode) SetURL(rawURL string) error {
	n.mu.Lock()
	defer n.mu.Unlock()

	if rawURL == "" {
		n.url = func(value primitive.Value) (string, error) {
			v := &url.URL{Scheme: "https"}

			if rawURL, ok := primitive.Pick[string](value, "url"); ok {
				var err error
				if v, err = url.Parse(rawURL); err != nil {
					return "", err
				}
			}

			if s, ok := primitive.Pick[string](value, "scheme"); ok {
				v.Scheme = s
			}
			if h, ok := primitive.Pick[string](value, "host"); ok {
				v.Host = h
			}
			if p, ok := primitive.Pick[string](value, "path"); ok {
				v.Path = p
			}

			return v.String(), nil
		}
		return nil
	}

	transform, err := language.CompileTransformWithPrimitive(rawURL, n.lang)
	if err != nil {
		return err
	}

	n.url = func(value primitive.Value) (string, error) {
		if v, err := transform(value); err != nil {
			return "", err
		} else {
			return fmt.Sprintf("%v", v.Interface()), nil
		}
	}
	return nil
}

// SetQuery sets the query parameters for the HTTP request.
func (n *HTTPClientNode) SetQuery(query string) error {
	n.mu.Lock()
	defer n.mu.Unlock()

	if query == "" {
		n.query = func(value primitive.Value) (url.Values, error) {
			if query, ok := primitive.Pick[map[string][]string](value, "query"); ok {
				return query, nil
			}
			if rawURL, ok := primitive.Pick[string](value, "url"); ok {
				if v, err := url.Parse(rawURL); err != nil {
					return nil, err
				} else {
					return v.Query(), nil
				}
			}
			return nil, nil
		}
		return nil
	}

	transform, err := language.CompileTransformWithPrimitive(query, n.lang)
	if err != nil {
		return err
	}

	n.query = func(value primitive.Value) (url.Values, error) {
		var encoded url.Values
		if v, err := transform(value); err != nil {
			return nil, err
		} else if err := primitive.Unmarshal(v, &encoded); err != nil {
			return nil, err
		} else {
			return encoded, nil
		}
	}
	return nil
}

// SetBody sets the body of the HTTP request.
func (n *HTTPClientNode) SetHeader(header string) error {
	n.mu.Lock()
	defer n.mu.Unlock()

	if header == "" {
		n.header = func(value primitive.Value) (http.Header, error) {
			if header, ok := primitive.Pick[map[string][]string](value, "header"); ok {
				return header, nil
			}
			return nil, nil
		}
		return nil
	}

	transform, err := language.CompileTransformWithPrimitive(header, n.lang)
	if err != nil {
		return err
	}

	n.header = func(value primitive.Value) (http.Header, error) {
		var encoded http.Header
		if v, err := transform(value); err != nil {
			return nil, err
		} else if err := primitive.Unmarshal(v, &encoded); err != nil {
			return nil, err
		} else {
			return encoded, nil
		}
	}
	return nil
}

// SetTimeout sets the timeout duration for the HTTP request.
func (n *HTTPClientNode) SetBody(body string) error {
	n.mu.Lock()
	defer n.mu.Unlock()

	if body == "" {
		n.body = func(value primitive.Value) (primitive.Value, error) {
			if body, ok := primitive.Pick[primitive.Value](value, "body"); ok {
				return body, nil
			}
			return nil, nil
		}
		return nil
	}

	transform, err := language.CompileTransformWithPrimitive(body, n.lang)
	if err != nil {
		return err
	}

	n.body = transform
	return nil
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

	ctx := proc.Context()
	if n.timeout != 0 {
		var cancel func()
		ctx, cancel = context.WithTimeout(ctx, n.timeout)
		defer cancel()
	}

	inPayload := inPck.Payload()

	req, err := n.request(inPayload)
	if err != nil {
		return nil, packet.WithError(err, inPck)
	}

	contentType := req.Header.Get(HeaderContentType)
	contentEncoding := req.Header.Get(HeaderContentEncoding)

	b, err := MarshalMIME(req.Body, &contentType)
	if err != nil {
		return nil, packet.WithError(err, inPck)
	}
	b, err = Compress(b, contentEncoding)
	if err != nil {
		return nil, packet.WithError(err, inPck)
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
		return nil, packet.WithError(err, inPck)
	}
	r = r.WithContext(ctx)

	client := &http.Client{}

	w, err := client.Do(r)
	if err != nil {
		return nil, packet.WithError(err, inPck)
	}
	defer w.Body.Close()

	res, err := n.response(w)
	if err != nil {
		return nil, packet.WithError(err, inPck)
	}

	outPayload, err := primitive.MarshalBinary(res)
	if err != nil {
		return nil, packet.WithError(err, inPck)
	}
	return packet.New(outPayload), nil
}

func (n *HTTPClientNode) request(raw primitive.Value) (*HTTPPayload, error) {
	method, err := n.method(raw)
	if err != nil {
		return nil, err
	}
	rawURL, err := n.url(raw)
	if err != nil {
		return nil, err
	}
	query, err := n.query(raw)
	if err != nil {
		return nil, err
	}
	header, err := n.header(raw)
	if err != nil {
		return nil, err
	}
	body, err := n.body(raw)
	if err != nil {
		return nil, err
	}

	if header == nil {
		header = make(http.Header)
	}

	u, err := url.Parse(rawURL)
	if err != nil {
		return nil, err
	}

	return &HTTPPayload{
		Method: method,
		Scheme: u.Scheme,
		Host:   u.Host,
		Path:   u.Path,
		Query:  query,
		Header: header,
		Body:   body,
	}, nil
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

func NewHTTPClientNodeCodec() scheme.Codec {
	return scheme.CodecWithType(func(spec *HTTPClientNodeSpec) (node.Node, error) {
		n := NewHTTPClientNode()

		n.SetLanguage(spec.Lang)
		if err := n.SetMethod(spec.Method); err != nil {
			return nil, err
		}
		if err := n.SetURL(spec.URL); err != nil {
			return nil, err
		}
		if err := n.SetQuery(spec.Query); err != nil {
			return nil, err
		}
		if err := n.SetHeader(spec.Header); err != nil {
			return nil, err
		}
		if err := n.SetBody(spec.Body); err != nil {
			return nil, err
		}

		return n, nil
	})
}
