package network

import (
	"io"
	"net"
	"net/http"
	"net/url"
	"strconv"
	"sync"

	"github.com/siyul-park/uniflow/pkg/node"
	"github.com/siyul-park/uniflow/pkg/packet"
	"github.com/siyul-park/uniflow/pkg/port"
	"github.com/siyul-park/uniflow/pkg/primitive"
	"github.com/siyul-park/uniflow/pkg/process"
	"github.com/siyul-park/uniflow/pkg/scheme"
)

// HTTPNode represents a Node for handling HTTP requests.
type HTTPNode struct {
	server   *http.Server
	listener net.Listener
	inPort   *port.InPort
	outPort  *port.OutPort
	errPort  *port.OutPort
	mu       sync.RWMutex
}

// HTTPNodeSpec holds the specifications for creating a HTTPNode.
type HTTPNodeSpec struct {
	scheme.SpecMeta `map:",inline"`
	Address         string `map:"address"`
}

// HTTPPayload is the payload structure for HTTP requests and responses.
type HTTPPayload struct {
	Method string          `map:"method,omitempty"`
	Scheme string          `map:"scheme,omitempty"`
	Host   string          `map:"host,omitempty"`
	Path   string          `map:"path,omitempty"`
	Query  url.Values      `map:"query,omitempty"`
	Proto  string          `map:"proto,omitempty"`
	Header http.Header     `map:"header,omitempty"`
	Body   primitive.Value `map:"body,omitempty"`
	Status int             `map:"status"`
}

const KindHTTP = "http"

const KeyHTTPRequest = "http.Request"
const KeyHTTPResponseWriter = "http.ResponseWriter"

var _ node.Node = (*HTTPNode)(nil)
var _ http.Handler = (*HTTPNode)(nil)

// NewHTTPPayload creates a new HTTPPayload with the given HTTP status code and optional body.
func NewHTTPPayload(status int, body ...primitive.Value) *HTTPPayload {
	if len(body) == 0 {
		body = []primitive.Value{primitive.String(http.StatusText(status))}
	}
	return &HTTPPayload{
		Header: http.Header{},
		Body:   body[0],
		Status: status,
	}
}

// NewHTTPNode creates a new HTTPNode with the specified address.
func NewHTTPNode(address string) *HTTPNode {
	n := &HTTPNode{
		inPort:  port.NewIn(),
		outPort: port.NewOut(),
		errPort: port.NewOut(),
	}

	n.inPort.AddHandler(port.HandlerFunc(n.forward))
	n.outPort.AddHandler(port.HandlerFunc(n.backward))
	n.outPort.AddHandler(port.HandlerFunc(n.catch))

	s := new(http.Server)
	s.Addr = address
	s.Handler = n
	n.server = s

	return n
}

func (n *HTTPNode) In(name string) *port.InPort {
	n.mu.RLock()
	defer n.mu.RUnlock()

	switch name {
	case node.PortIn:
		return n.inPort
	default:
	}

	return nil
}

func (n *HTTPNode) Out(name string) *port.OutPort {
	n.mu.RLock()
	defer n.mu.RUnlock()

	switch name {
	case node.PortOut:
		return n.outPort
	case node.PortErr:
		return n.errPort
	default:
	}

	return nil
}

// Address returns the listener address if available.
func (n *HTTPNode) Address() net.Addr {
	n.mu.RLock()
	defer n.mu.RUnlock()

	if n.listener == nil {
		return nil
	}
	return n.listener.Addr()
}

// Listen starts the HTTP server.
func (n *HTTPNode) Listen() error {
	if err := func() error {
		n.mu.Lock()
		defer n.mu.Unlock()

		if n.listener != nil {
			return nil
		}
		if l, err := newTCPKeepAliveListener(n.server.Addr, "tcp"); err != nil {
			return err
		} else {
			n.listener = l
		}
		return nil
	}(); err != nil {
		return err
	}

	go n.server.Serve(n.listener)

	return nil
}

// ServeHTTP handles HTTP requests.
func (n *HTTPNode) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	n.mu.RLock()
	defer n.mu.RUnlock()

	proc := process.New()

	proc.Heap().Store(KeyHTTPResponseWriter, w)
	proc.Heap().Store(KeyHTTPRequest, r)

	outWriter := n.outPort.Open(proc)

	if req, err := n.read(r); err != nil {
		n.throw(proc, packet.WithError(err, nil))
	} else if outPayload, err := primitive.MarshalBinary(req); err != nil {
		n.throw(proc, packet.WithError(err, nil))
	} else if outWriter.Links() > 0 {
		outPck := packet.New(outPayload)
		proc.Stack().Add(nil, outPck)
		outWriter.Write(outPck)

		<-proc.Stack().Done(outPck)
	}

	go proc.Close()
}

// Close closes all ports and stops the HTTP server.
func (n *HTTPNode) Close() error {
	n.mu.Lock()
	defer n.mu.Unlock()

	n.inPort.Close()
	n.outPort.Close()
	n.errPort.Close()

	return n.server.Close()
}

func (n *HTTPNode) forward(proc *process.Process) {
	n.mu.RLock()
	defer n.mu.RUnlock()

	inReader := n.inPort.Open(proc)

	for {
		inPck, ok := <-inReader.Read()
		if !ok {
			return
		}

		proc.Stack().Clear(inPck)
		n.receive(proc, inPck)
	}
}

func (n *HTTPNode) backward(proc *process.Process) {
	n.mu.RLock()
	defer n.mu.RUnlock()

	outWriter := n.outPort.Open(proc)

	for {
		backPck, ok := <-outWriter.Receive()
		if !ok {
			return
		}

		if _, ok := packet.AsError(backPck); ok {
			n.throw(proc, backPck)
		} else {
			n.receive(proc, backPck)
		}
	}
}

func (n *HTTPNode) catch(proc *process.Process) {
	n.mu.RLock()
	defer n.mu.RUnlock()

	errWriter := n.errPort.Open(proc)

	for {
		backPck, ok := <-errWriter.Receive()
		if !ok {
			return
		}

		n.receive(proc, backPck)
	}
}

func (n *HTTPNode) throw(proc *process.Process, errPck *packet.Packet) {
	errWriter := n.errPort.Open(proc)

	if errWriter.Links() > 0 {
		errWriter.Write(errPck)
	} else {
		n.receive(proc, errPck)
	}
}

func (n *HTTPNode) receive(proc *process.Process, backPck *packet.Packet) {
	var res *HTTPPayload
	if err, ok := packet.AsError(backPck); ok {
		res = NewHTTPPayload(http.StatusInternalServerError)
		proc.SetErr(err)
	} else if err := primitive.Unmarshal(backPck.Payload(), &res); err != nil {
		res.Body = backPck.Payload()
	}

	if r, ok := proc.Heap().Load(KeyHTTPRequest).(*http.Request); ok {
		acceptEncoding := r.Header.Get(HeaderAcceptEncoding)
		accept := r.Header.Get(HeaderAccept)

		contentEncoding := Negotiate(acceptEncoding, []string{EncodingIdentity, EncodingGzip, EncodingDeflate, EncodingBr})
		contentType := Negotiate(accept, []string{ApplicationJSON, ApplicationForm, ApplicationOctetStream, TextPlain, MultipartFormData})

		negotiate := func(payload *HTTPPayload) {
			if payload == nil {
				return
			}
			if payload.Header == nil {
				payload.Header = http.Header{}
			}
			if payload.Header.Get(HeaderContentEncoding) == "" {
				payload.Header.Set(HeaderContentEncoding, contentEncoding)
			}
			if payload.Header.Get(HeaderContentType) == "" {
				payload.Header.Set(HeaderContentType, contentType)
			}
		}

		negotiate(res)

		if w, ok := proc.Heap().LoadAndDelete(KeyHTTPResponseWriter).(http.ResponseWriter); ok {
			if err := n.write(w, res); err != nil {
				proc.SetErr(err)

				res = NewHTTPPayload(http.StatusInternalServerError)
				negotiate(res)

				_ = n.write(w, res)
			}
		}

		proc.Stack().Clear(backPck)
	}
}

func (n *HTTPNode) read(r *http.Request) (*HTTPPayload, error) {
	contentType := r.Header.Get(HeaderContentType)
	contentEncoding := r.Header.Get(HeaderContentEncoding)

	if b, err := io.ReadAll(r.Body); err != nil {
		return nil, err
	} else if b, err := Decompress(b, contentEncoding); err != nil {
		return nil, err
	} else if b, err := UnmarshalMIME(b, &contentType); err != nil {
		return nil, err
	} else {
		r.Header.Set(HeaderContentType, contentType)
		return &HTTPPayload{
			Method: r.Method,
			Scheme: r.URL.Scheme,
			Host:   r.Host,
			Path:   r.URL.Path,
			Query:  r.URL.Query(),
			Proto:  r.Proto,
			Header: r.Header,
			Body:   b,
		}, nil
	}
}

func (n *HTTPNode) write(w http.ResponseWriter, res *HTTPPayload) error {
	if res == nil {
		return nil
	}

	contentType := res.Header.Get(HeaderContentType)
	contentEncoding := res.Header.Get(HeaderContentEncoding)

	b, err := MarshalMIME(res.Body, &contentType)
	if err != nil {
		return err
	}
	b, err = Compress(b, contentEncoding)
	if err != nil {
		return err
	}

	if res.Header == nil {
		res.Header = http.Header{}
	}
	res.Header.Set(HeaderContentType, contentType)
	for key := range w.Header() {
		w.Header().Del(key)
	}
	for key, headers := range res.Header {
		if !IsResponseHeader(key) {
			continue
		}
		for _, header := range headers {
			w.Header().Add(key, header)
		}
	}
	w.Header().Set(HeaderContentLength, strconv.Itoa(len(b)))
	w.Header().Set(HeaderContentType, contentType)

	status := res.Status
	if status == 0 {
		status = http.StatusOK
	}
	w.WriteHeader(status)

	if _, err := w.Write(b); err != nil {
		return err
	}
	if f, ok := w.(http.Flusher); ok {
		f.Flush()
	}
	return nil
}

// NewHTTPNodeCodec creates a new codec for HTTPNodeSpec.
func NewHTTPNodeCodec() scheme.Codec {
	return scheme.CodecWithType[*HTTPNodeSpec](func(spec *HTTPNodeSpec) (node.Node, error) {
		return NewHTTPNode(spec.Address), nil
	})
}
