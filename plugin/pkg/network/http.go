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
	ioPort   *port.Port
	inPort   *port.Port
	outPort  *port.Port
	errPort  *port.Port
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

const KeyHTTPResponseWriter = "http.ResponseWriter"

var _ node.Node = (*HTTPNode)(nil)
var _ http.Handler = (*HTTPNode)(nil)

// NewHTTPNodeCodec creates a new codec for HTTPNodeSpec.
func NewHTTPNodeCodec() scheme.Codec {
	return scheme.CodecWithType[*HTTPNodeSpec](func(spec *HTTPNodeSpec) (node.Node, error) {
		return NewHTTPNode(spec.Address), nil
	})
}

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
		ioPort:  port.New(),
		inPort:  port.New(),
		outPort: port.New(),
		errPort: port.New(),
	}

	s := new(http.Server)
	s.Addr = address
	s.Handler = n
	n.server = s

	return n
}

// Port returns the specified port.
func (n *HTTPNode) Port(name string) (*port.Port, bool) {
	n.mu.RLock()
	defer n.mu.RUnlock()

	switch name {
	case node.PortIO:
		return n.ioPort, true
	case node.PortIn:
		return n.inPort, true
	case node.PortOut:
		return n.outPort, true
	case node.PortErr:
		return n.errPort, true
	default:
	}

	return nil, false
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
	proc.Share().Store(KeyHTTPResponseWriter, w)

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
	write := func(payload *HTTPPayload) error {
		negotiate(payload)
		return n.write(w, payload)
	}
	writeError := func(err error) error {
		return write(n.newErrorPayload(proc, err))
	}

	var req *HTTPPayload
	var res *HTTPPayload
	var err error
	if req, err = n.read(r); err != nil {
		_ = writeError(err)
	} else if res, err = n.action(proc, req); err != nil {
		_ = writeError(err)
	} else if err = write(res); err != nil {
		_ = writeError(err)
	}

	proc.Stack().Wait()
	proc.Exit(err)
}

// Close closes all ports and stops the HTTP server.
func (n *HTTPNode) Close() error {
	n.mu.Lock()
	defer n.mu.Unlock()

	n.ioPort.Close()
	n.inPort.Close()
	n.outPort.Close()
	n.errPort.Close()

	return n.server.Close()
}

func (n *HTTPNode) action(proc *process.Process, req *HTTPPayload) (*HTTPPayload, error) {
	ioStream := n.ioPort.Open(proc)
	inStream := n.inPort.Open(proc)
	outStream := n.outPort.Open(proc)

	if ioStream.Links()+outStream.Links() == 0 {
		return nil, nil
	}

	outPayload, err := primitive.MarshalBinary(req)
	if err != nil {
		return nil, err
	}

	ioPck := packet.New(outPayload)
	outPck := packet.New(outPayload)

	if ioStream.Links() > 0 {
		proc.Stack().Push(ioPck.ID(), ioStream.ID())
		ioStream.Send(ioPck)
	}
	if outStream.Links() > 0 {
		proc.Stack().Push(outPck.ID(), outStream.ID())
		outStream.Send(outPck)
	}

	var inPck *packet.Packet
	var ok bool
	select {
	case inPck, ok = <-ioStream.Receive():
		if ok {
			_, ok = proc.Stack().Pop(inPck.ID(), ioStream.ID())
		}
	case inPck, ok = <-inStream.Receive():
	case inPck, ok = <-outStream.Receive():
		if ok {
			_, ok = proc.Stack().Pop(inPck.ID(), outStream.ID())
		}
	}

	proc.Stack().Clear(ioPck.ID())
	proc.Stack().Clear(inPck.ID())
	proc.Stack().Clear(outPck.ID())

	if !ok {
		return NewHTTPPayload(http.StatusInternalServerError), nil
	}

	if err, ok := packet.AsError(inPck); ok {
		return nil, err
	}

	inPayload := inPck.Payload()
	if inPayload == nil {
		return nil, nil
	}

	var res *HTTPPayload
	if err := primitive.Unmarshal(inPayload, &res); err != nil {
		res.Body = inPayload
	}

	return res, nil
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

func (n *HTTPNode) newErrorPayload(proc *process.Process, err error) *HTTPPayload {
	errStream := n.errPort.Open(proc)
	if errStream.Links() == 0 {
		return NewHTTPPayload(http.StatusInternalServerError)
	}

	errPck := packet.WithError(err, nil)
	errStream.Send(errPck)

	inPck, ok := <-errStream.Receive()
	if !ok {
		return NewHTTPPayload(http.StatusInternalServerError)
	}
	if _, ok := packet.AsError(inPck); ok {
		return NewHTTPPayload(http.StatusInternalServerError)
	}

	inPayload := inPck.Payload()
	if inPayload == nil {
		return nil
	}

	var res *HTTPPayload
	if err := primitive.Unmarshal(inPayload, &res); err != nil {
		res.Body = inPayload
	}
	return res
}
