package network

import (
	"io"
	"net"
	"net/http"
	"strconv"
	"sync"

	"github.com/siyul-park/uniflow/pkg/node"
	"github.com/siyul-park/uniflow/pkg/packet"
	"github.com/siyul-park/uniflow/pkg/port"
	"github.com/siyul-park/uniflow/pkg/primitive"
	"github.com/siyul-park/uniflow/pkg/process"
	"github.com/siyul-park/uniflow/pkg/scheme"
	"github.com/siyul-park/uniflow/pkg/transaction"
)

// HTTPServerNode represents a Node for handling HTTP requests.
type HTTPServerNode struct {
	server   *http.Server
	listener net.Listener
	inPort   *port.InPort
	outPort  *port.OutPort
	errPort  *port.OutPort
	mu       sync.RWMutex
}

// HTTPServerNodeSpec holds the specifications for creating a HTTPServerNode.
type HTTPServerNodeSpec struct {
	scheme.SpecMeta `map:",inline"`
	Address         string `map:"address"`
}

const KindHTTPServer = "http/server"

var _ node.Node = (*HTTPServerNode)(nil)
var _ http.Handler = (*HTTPServerNode)(nil)

// NewHTTPServerNode creates a new HTTPServerNode with the specified address.
func NewHTTPServerNode(address string) *HTTPServerNode {
	n := &HTTPServerNode{
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

// In returns the input port with the specified name.
func (n *HTTPServerNode) In(name string) *port.InPort {
	n.mu.RLock()
	defer n.mu.RUnlock()

	switch name {
	case node.PortIn:
		return n.inPort
	default:
	}

	return nil
}

// Out returns the output port with the specified name.
func (n *HTTPServerNode) Out(name string) *port.OutPort {
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
func (n *HTTPServerNode) Address() net.Addr {
	n.mu.RLock()
	defer n.mu.RUnlock()

	if n.listener == nil {
		return nil
	}
	return n.listener.Addr()
}

// Listen starts the HTTP server.
func (n *HTTPServerNode) Listen() error {
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
func (n *HTTPServerNode) ServeHTTP(w http.ResponseWriter, r *http.Request) {
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
		tx := transaction.New()

		proc.Stack().Add(nil, outPck)
		proc.SetTransaction(outPck, tx)

		if !outWriter.Write(outPck) {
			proc.Stack().Clear(outPck)
		}

		<-proc.Stack().Done(outPck)
		_ = tx.Commit()
	}

	go proc.Close()
}

// Close closes all ports and stops the HTTP server.
func (n *HTTPServerNode) Close() error {
	n.mu.Lock()
	defer n.mu.Unlock()

	n.inPort.Close()
	n.outPort.Close()
	n.errPort.Close()

	return n.server.Close()
}

func (n *HTTPServerNode) forward(proc *process.Process) {
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

func (n *HTTPServerNode) backward(proc *process.Process) {
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

func (n *HTTPServerNode) catch(proc *process.Process) {
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

func (n *HTTPServerNode) throw(proc *process.Process, errPck *packet.Packet) {
	errWriter := n.errPort.Open(proc)

	if !errWriter.Write(errPck) {
		n.receive(proc, errPck)
	}
}

func (n *HTTPServerNode) receive(proc *process.Process, backPck *packet.Packet) {
	var res *HTTPPayload
	if _, ok := packet.AsError(backPck); ok {
		res = NewHTTPPayload(http.StatusInternalServerError)
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
				res = NewHTTPPayload(http.StatusInternalServerError)
				negotiate(res)

				_ = n.write(w, res)
			}
		}
	}

	tx := proc.Transaction(backPck)
	if res.Status < 400 {
		tx.Commit()
	} else {
		tx.Rollback()
	}

	proc.Stack().Clear(backPck)
}

func (n *HTTPServerNode) read(r *http.Request) (*HTTPPayload, error) {
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

func (n *HTTPServerNode) write(w http.ResponseWriter, res *HTTPPayload) error {
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

// NewHTTPServerNodeCodec creates a new codec for HTTPServerNodeSpec.
func NewHTTPServerNodeCodec() scheme.Codec {
	return scheme.CodecWithType(func(spec *HTTPServerNodeSpec) (node.Node, error) {
		return NewHTTPServerNode(spec.Address), nil
	})
}
