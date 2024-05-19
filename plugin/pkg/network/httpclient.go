package network

import (
	"bytes"
	"context"
	"io"
	"net/http"
	"net/url"
	"reflect"
	"strconv"
	"sync"
	"time"

	"github.com/pkg/errors"
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
	method  func(any) (string, error)
	url     func(any) (string, error)
	query   func(any) (url.Values, error)
	header  func(any) (http.Header, error)
	body    func(any) (primitive.Value, error)
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
		n.method = func(input any) (string, error) {
			value := reflect.ValueOf(input)
			if value.Kind() == reflect.Map {
				if e := value.MapIndex(reflect.ValueOf("method")); e.IsValid() {
					if e, ok := e.Interface().(string); ok {
						return e, nil
					}
				}
			}
			return http.MethodGet, nil
		}
		return nil
	}

	lang := n.lang
	transform, err := language.CompileTransform(method, &lang)
	if err != nil {
		return err
	}

	n.method = func(input any) (string, error) {
		if output, err := transform(input); err != nil {
			return "", err
		} else if v, ok := output.(string); ok {
			return v, nil
		}
		return "", errors.WithStack(packet.ErrInvalidPacket)
	}
	return nil
}

// SetURL sets the target URL for the HTTP request.
func (n *HTTPClientNode) SetURL(rawURL string) error {
	n.mu.Lock()
	defer n.mu.Unlock()

	if rawURL == "" {
		n.url = func(input any) (string, error) {
			v := &url.URL{Scheme: "https"}

			value := reflect.ValueOf(input)
			if value.Kind() == reflect.Map {
				if e := value.MapIndex(reflect.ValueOf("url")); e.IsValid() {
					if e, ok := e.Interface().(string); ok {
						var err error
						if v, err = url.Parse(e); err != nil {
							return "", err
						}
					}
				}

				if e := value.MapIndex(reflect.ValueOf("scheme")); e.IsValid() {
					if e, ok := e.Interface().(string); ok {
						v.Scheme = e
					}
				}
				if e := value.MapIndex(reflect.ValueOf("host")); e.IsValid() {
					if e, ok := e.Interface().(string); ok {
						v.Host = e
					}
				}
				if e := value.MapIndex(reflect.ValueOf("path")); e.IsValid() {
					if e, ok := e.Interface().(string); ok {
						v.Path = e
					}
				}
			}

			return v.String(), nil
		}
		return nil
	}

	lang := n.lang
	transform, err := language.CompileTransform(rawURL, &lang)
	if err != nil {
		return err
	}

	n.url = func(input any) (string, error) {
		if output, err := transform(input); err != nil {
			return "", err
		} else if v, ok := output.(string); ok {
			return v, nil
		}
		return "", errors.WithStack(packet.ErrInvalidPacket)
	}
	return nil
}

// SetQuery sets the query parameters for the HTTP request.
func (n *HTTPClientNode) SetQuery(query string) error {
	n.mu.Lock()
	defer n.mu.Unlock()

	if query == "" {
		n.query = func(input any) (url.Values, error) {
			value := reflect.ValueOf(input)
			if value.Kind() == reflect.Map {
				if e := value.MapIndex(reflect.ValueOf("query")); e.IsValid() {
					var v url.Values
					if e, err := primitive.MarshalText(e.Interface()); err != nil {
						return nil, err
					} else if err := primitive.Unmarshal(e, &v); err != nil {
						return nil, err
					}
					return v, nil
				}

				if e := value.MapIndex(reflect.ValueOf("url")); e.IsValid() {
					if e, ok := e.Interface().(string); ok {
						if v, err := url.Parse(e); err != nil {
							return nil, err
						} else {
							return v.Query(), nil
						}
					}
				}
			}
			return nil, nil
		}
		return nil
	}

	lang := n.lang
	transform, err := language.CompileTransform(query, &lang)
	if err != nil {
		return err
	}

	n.query = func(input any) (url.Values, error) {
		output, err := transform(input)
		if err != nil {
			return nil, err
		}

		var v url.Values
		if e, err := primitive.MarshalText(output); err != nil {
			return nil, err
		} else if err := primitive.Unmarshal(e, &v); err != nil {
			return nil, err
		}
		return v, nil
	}
	return nil
}

// SetBody sets the body of the HTTP request.
func (n *HTTPClientNode) SetHeader(header string) error {
	n.mu.Lock()
	defer n.mu.Unlock()

	if header == "" {
		n.header = func(input any) (http.Header, error) {
			value := reflect.ValueOf(input)
			if value.Kind() == reflect.Map {
				if e := value.MapIndex(reflect.ValueOf("header")); e.IsValid() {
					var v http.Header
					if e, err := primitive.MarshalText(e); err != nil {
						return nil, err
					} else if err := primitive.Unmarshal(e, &v); err != nil {
						return nil, err
					}
					return v, nil
				}
			}
			return nil, nil
		}
		return nil
	}

	lang := n.lang
	transform, err := language.CompileTransform(header, &lang)
	if err != nil {
		return err
	}

	n.header = func(input any) (http.Header, error) {
		output, err := transform(input)
		if err != nil {
			return nil, err
		}

		var v http.Header
		if e, err := primitive.MarshalText(output); err != nil {
			return nil, err
		} else if err := primitive.Unmarshal(e, &v); err != nil {
			return nil, err
		}
		return v, nil
	}
	return nil
}

// SetTimeout sets the timeout duration for the HTTP request.
func (n *HTTPClientNode) SetBody(body string) error {
	n.mu.Lock()
	defer n.mu.Unlock()

	if body == "" {
		n.body = func(input any) (primitive.Value, error) {
			value := reflect.ValueOf(input)
			if value.Kind() == reflect.Map {
				if e := value.MapIndex(reflect.ValueOf("body")); e.IsValid() {
					return primitive.MarshalText(e)
				}
			}
			return nil, nil
		}
		return nil
	}

	lang := n.lang
	transform, err := language.CompileTransform(body, &lang)
	if err != nil {
		return err
	}

	n.body = func(input any) (primitive.Value, error) {
		output, err := transform(input)
		if err != nil {
			return nil, err
		}
		return primitive.MarshalText(output)
	}
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
	input := primitive.Interface(inPayload)

	req, err := n.request(input)
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

	outPayload, err := primitive.MarshalText(res)
	if err != nil {
		return nil, packet.WithError(err, inPck)
	}
	return packet.New(outPayload), nil
}

func (n *HTTPClientNode) request(input any) (*HTTPPayload, error) {
	method, err := n.method(input)
	if err != nil {
		return nil, err
	}
	rawURL, err := n.url(input)
	if err != nil {
		return nil, err
	}
	query, err := n.query(input)
	if err != nil {
		return nil, err
	}
	header, err := n.header(input)
	if err != nil {
		return nil, err
	}
	body, err := n.body(input)
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
