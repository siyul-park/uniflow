package network

import (
	"bytes"
	"context"
	"fmt"
	"github.com/siyul-park/uniflow/pkg/node"
	"github.com/siyul-park/uniflow/pkg/packet"
	"github.com/siyul-park/uniflow/pkg/primitive"
	"github.com/siyul-park/uniflow/pkg/process"
	"github.com/siyul-park/uniflow/pkg/scheme"
	"github.com/siyul-park/uniflow/plugin/internal/language"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"sync"
	"time"
)

type CHTTPNode struct {
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

type CHTTPNodeSpec struct {
	scheme.SpecMeta `map:",inline"`
	Lang            string              `map:"lang,omitempty"`
	Method          string              `map:"method,omitempty"`
	URL             string              `map:"url,omitempty"`
	Query           map[string][]string `map:"query,omitempty"`
	Header          map[string][]string `map:"header,omitempty"`
	Body            primitive.Value     `map:"body,omitempty"`
}

const KindCHTTP = "chttp"

func NewCHTTPNode() *CHTTPNode {
	n := &CHTTPNode{}
	n.OneToOneNode = node.NewOneToOneNode(n.action)

	_ = n.SetMethod("")
	_ = n.SetURL("")
	_ = n.SetQuery(nil)
	_ = n.SetHeader(nil)
	_ = n.SetBody(nil)

	return n
}

func (n *CHTTPNode) SetLanguage(lang string) {
	n.mu.Lock()
	defer n.mu.Unlock()

	n.lang = lang
}

func (n *CHTTPNode) SetMethod(method string) error {
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

	transform, err := n.compileText(method)
	if err != nil {
		return err
	}
	n.method = transform
	return nil
}

func (n *CHTTPNode) SetURL(value string) error {
	n.mu.Lock()
	defer n.mu.Unlock()

	if value == "" {
		n.url = func(value primitive.Value) (string, error) {
			v := &url.URL{
				Scheme: "https",
				Path:   "",
			}

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

	transform, err := n.compileText(value)
	if err != nil {
		return err
	}
	n.url = transform
	return nil
}

func (n *CHTTPNode) SetQuery(query map[string][]string) error {
	n.mu.Lock()
	defer n.mu.Unlock()

	if query == nil {
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

	transforms := make(map[string][]func(primitive.Value) (string, error), len(query))
	for k, values := range query {
		for _, v := range values {
			transform, err := n.compileText(v)
			if err != nil {
				return err
			}
			transforms[k] = append(transforms[k], transform)
		}
	}

	n.query = func(value primitive.Value) (url.Values, error) {
		outputs := make(url.Values, len(transforms))
		for k, transforms := range transforms {
			for _, transform := range transforms {
				output, err := transform(value)
				if err != nil {
					return nil, err
				}
				outputs[k] = append(outputs[k], output)
			}
		}
		return outputs, nil
	}
	return nil
}

func (n *CHTTPNode) SetHeader(header map[string][]string) error {
	n.mu.Lock()
	defer n.mu.Unlock()

	if header == nil {
		n.header = func(value primitive.Value) (http.Header, error) {
			if header, ok := primitive.Pick[map[string][]string](value, "header"); ok {
				return header, nil
			}
			return nil, nil
		}
		return nil
	}

	transforms := make(map[string][]func(primitive.Value) (string, error), len(header))
	for k, values := range header {
		for _, v := range values {
			transform, err := n.compileText(v)
			if err != nil {
				return err
			}
			transforms[k] = append(transforms[k], transform)
		}
	}

	n.header = func(value primitive.Value) (http.Header, error) {
		outputs := make(http.Header, len(transforms))
		for k, transforms := range transforms {
			for _, transform := range transforms {
				output, err := transform(value)
				if err != nil {
					return nil, err
				}
				outputs[k] = append(outputs[k], output)
			}
		}
		return outputs, nil
	}
	return nil
}

func (n *CHTTPNode) SetBody(body primitive.Value) error {
	n.mu.Lock()
	defer n.mu.Unlock()

	if body == nil {
		n.body = func(value primitive.Value) (primitive.Value, error) {
			if body, ok := primitive.Pick[primitive.Value](value, "body"); ok {
				return body, nil
			}
			return nil, nil
		}
		return nil
	}

	transform, err := n.compileValue(body)
	if err != nil {
		return err
	}
	n.body = transform
	return nil
}

func (n *CHTTPNode) SetTimeout(timeout time.Duration) {
	n.mu.Lock()
	defer n.mu.Unlock()

	n.timeout = timeout
}

func (n *CHTTPNode) action(proc *process.Process, inPck *packet.Packet) (*packet.Packet, *packet.Packet) {
	n.mu.RLock()
	defer n.mu.RUnlock()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	if n.timeout != 0 {
		ctx, cancel = context.WithTimeout(ctx, n.timeout)
		defer cancel()
	}

	go func() {
		<-proc.Done()
		cancel()
	}()

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
	defer func() { _ = w.Body.Close() }()

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

func (n *CHTTPNode) request(raw primitive.Value) (*HTTPPayload, error) {
	method, err := n.method(raw)
	if err != nil {
		return nil, err
	}
	rawURL, err := n.url(raw)
	if err != nil {
		return nil, err
	}
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

func (n *CHTTPNode) response(w *http.Response) (*HTTPPayload, error) {
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

func (n *CHTTPNode) compileValue(value primitive.Value) (func(primitive.Value) (primitive.Value, error), error) {
	switch v := value.(type) {
	case *primitive.Map:
		transforms := make([]func(primitive.Value) (primitive.Value, error), 0, v.Len())
		for _, k := range v.Keys() {
			transform, err := n.compileValue(v.GetOr(k, nil))
			if err != nil {
				return nil, err
			}
			transforms = append(transforms, transform)
		}
		return func(value primitive.Value) (primitive.Value, error) {
			pairs := make([]primitive.Value, 0, v.Len()*2)
			for i, k := range v.Keys() {
				transform := transforms[i]

				v, err := transform(value)
				if err != nil {
					return nil, err
				}

				pairs = append(pairs, k)
				pairs = append(pairs, v)
			}
			return primitive.NewMap(pairs...), nil
		}, nil
	case *primitive.Slice:
		transforms := make([]func(primitive.Value) (primitive.Value, error), 0, v.Len())
		for _, v := range v.Values() {
			transform, err := n.compileValue(v)
			if err != nil {
				return nil, err
			}
			transforms = append(transforms, transform)
		}
		return func(value primitive.Value) (primitive.Value, error) {
			values := make([]primitive.Value, 0, v.Len()*2)
			for i, v := range v.Values() {
				transform := transforms[i]

				v, err := transform(v)
				if err != nil {
					return nil, err
				}

				values = append(values, v)
			}
			return primitive.NewSlice(values...), nil
		}, nil
	case primitive.String:
		transform, err := n.compile(v.String())
		if err != nil {
			return nil, err
		}
		return func(value primitive.Value) (primitive.Value, error) {
			if out, err := transform(value); err != nil {
				return nil, err
			} else {
				return primitive.MarshalBinary(out)
			}
		}, nil
	default:
		return func(value primitive.Value) (primitive.Value, error) {
			return v, nil
		}, nil
	}
}

func (n *CHTTPNode) compileText(code string) (func(primitive.Value) (string, error), error) {
	transform, err := n.compile(code)
	if err != nil {
		return nil, err
	}

	return func(value primitive.Value) (string, error) {
		if out, err := transform(value); err != nil {
			return "", err
		} else {
			return fmt.Sprintf("%v", out), nil
		}
	}, nil
}

func (n *CHTTPNode) compile(code string) (func(primitive.Value) (any, error), error) {
	lang := n.lang
	transform, err := language.CompileTransform(code, &lang)
	if err != nil {
		return nil, err
	}

	return func(value primitive.Value) (any, error) {
		var input any
		switch lang {
		case language.Typescript, language.Javascript, language.JSONata:
			input = primitive.Interface(value)
		}
		return transform(input)
	}, nil
}

func NewCHTTPNodeCodec() scheme.Codec {
	return scheme.CodecWithType[*CHTTPNodeSpec](func(spec *CHTTPNodeSpec) (node.Node, error) {
		n := NewCHTTPNode()
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
