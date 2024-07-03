package network

import (
	"bytes"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/textproto"
	"sync"

	"github.com/pkg/errors"
	"github.com/siyul-park/uniflow/ext/pkg/mime"
	"github.com/siyul-park/uniflow/pkg/node"
	"github.com/siyul-park/uniflow/pkg/object"
	"github.com/siyul-park/uniflow/pkg/packet"
	"github.com/siyul-park/uniflow/pkg/port"
	"github.com/siyul-park/uniflow/pkg/process"
	"github.com/siyul-park/uniflow/pkg/scheme"
	"github.com/siyul-park/uniflow/pkg/spec"
)

// HTTPListenNode represents a Node for handling HTTP requests.
type HTTPListenNode struct {
	server   *http.Server
	listener net.Listener
	outPort  *port.OutPort
	errPort  *port.OutPort
	mu       sync.RWMutex
}

// ListenNodeSpec holds the specifications for creating a ListenNode.
type ListenNodeSpec struct {
	spec.Meta `map:",inline"`
	Protocol  string `map:"protocol"`
	Port      int    `map:"port"`
}

const KindListen = "listen"

const KeyHTTPRequest = "http.Request"
const KeyHTTPResponseWriter = "http.ResponseWriter"

var _ node.Node = (*HTTPListenNode)(nil)
var _ http.Handler = (*HTTPListenNode)(nil)

// NewHTTPListenNode creates a new HTTPListenNode with the specified address.
func NewHTTPListenNode(address string) *HTTPListenNode {
	n := &HTTPListenNode{
		outPort: port.NewOut(),
		errPort: port.NewOut(),
	}

	s := new(http.Server)
	s.Addr = address
	s.Handler = n
	n.server = s

	return n
}

// In returns the input port with the specified name.
func (n *HTTPListenNode) In(name string) *port.InPort {
	return nil
}

// Out returns the output port with the specified name.
func (n *HTTPListenNode) Out(name string) *port.OutPort {
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
func (n *HTTPListenNode) Address() net.Addr {
	n.mu.RLock()
	defer n.mu.RUnlock()

	if n.listener == nil {
		return nil
	}
	return n.listener.Addr()
}

// Listen starts the HTTP server.
func (n *HTTPListenNode) Listen() error {
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

// Shutdown shuts down the HTTPListenNode by closing the server and its associated listener.
// It locks the mutex to ensure safe concurrent access to the server and listener.
// If an error occurs during the shutdown process, it returns the error.
func (n *HTTPListenNode) Shutdown() error {
	n.mu.Lock()
	defer n.mu.Unlock()

	if n.listener != nil {
		return nil
	}

	if err := n.server.Close(); err != nil {
		return err
	}
	s := new(http.Server)
	s.Addr = n.server.Addr
	n.server = s

	_ = n.listener.Close()
	n.listener = nil

	return nil
}

// ServeHTTP handles HTTP requests.
func (n *HTTPListenNode) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	n.mu.RLock()
	defer n.mu.RUnlock()

	proc := process.New()
	ctx := r.Context()

	go func() {
		<-ctx.Done()
		proc.Wait()
		proc.Exit(ctx.Err())
	}()

	proc.Data().Store(KeyHTTPResponseWriter, w)
	proc.Data().Store(KeyHTTPRequest, r)

	outWriter := n.outPort.Open(proc)
	errWriter := n.errPort.Open(proc)

	var outPck *packet.Packet
	var errPck *packet.Packet
	req, err := n.read(r)
	if err != nil {
		errPck = packet.New(object.NewError(err))
	} else if outPayload, err := object.MarshalText(req); err != nil {
		errPck = packet.New(object.NewError(err))
	} else {
		outPck = packet.New(outPayload)
	}

	var backPck *packet.Packet
	if errPck != nil {
		backPck = packet.Call(errWriter, errPck)
	} else {
		backPck = packet.Call(outWriter, outPck)
		if _, ok := backPck.Payload().(object.Error); ok {
			backPck = packet.CallOrFallback(errWriter, backPck, backPck)
		}
	}

	err = nil
	if backPck != packet.None {
		var res *HTTPPayload
		if _, ok := backPck.Payload().(object.Error); ok {
			res = NewHTTPPayload(http.StatusInternalServerError)
		} else if err := object.Unmarshal(backPck.Payload(), &res); err != nil {
			res.Body = backPck.Payload()
		}

		if res.Status >= 400 && res.Status < 600 {
			err = errors.New(http.StatusText(res.Status))
		}

		if w, ok := proc.Data().LoadAndDelete(KeyHTTPResponseWriter).(http.ResponseWriter); ok {
			n.negotiate(req, res)
			_ = n.write(w, res)
		}
	}

	go func() {
		proc.Wait()
		proc.Exit(err)
	}()
}

// Close closes all ports and stops the HTTP server.
func (n *HTTPListenNode) Close() error {
	n.mu.Lock()
	defer n.mu.Unlock()

	n.outPort.Close()
	n.errPort.Close()

	return n.server.Close()
}

func (n *HTTPListenNode) negotiate(req *HTTPPayload, res *HTTPPayload) {
	if res.Header == nil {
		res.Header = http.Header{}
	}
	if res.Header.Get(mime.HeaderContentEncoding) == "" {
		acceptEncoding := req.Header.Get(mime.HeaderAcceptEncoding)
		res.Header.Set(mime.HeaderContentEncoding, mime.Negotiate(acceptEncoding, []string{mime.EncodingIdentity, mime.EncodingGzip, mime.EncodingDeflate, mime.EncodingBr}))
	}
	if res.Header.Get(mime.HeaderContentType) == "" {
		accept := req.Header.Get(mime.HeaderAccept)
		res.Header.Set(mime.HeaderContentType, mime.Negotiate(accept, []string{mime.ApplicationJSON, mime.ApplicationFormURLEncoded, mime.ApplicationOctetStream, mime.TextPlain, mime.MultipartFormData}))
	}
}

func (n *HTTPListenNode) read(r *http.Request) (*HTTPPayload, error) {
	if body, err := mime.Decode(r.Body, textproto.MIMEHeader(r.Header)); err != nil {
		return nil, err
	} else {
		return &HTTPPayload{
			Method: r.Method,
			Scheme: r.URL.Scheme,
			Host:   r.Host,
			Path:   r.URL.Path,
			Query:  r.URL.Query(),
			Proto:  r.Proto,
			Header: r.Header,
			Body:   body,
		}, nil
	}
}

func (n *HTTPListenNode) write(w http.ResponseWriter, res *HTTPPayload) error {
	if res == nil {
		return nil
	}
	if res.Header == nil {
		res.Header = http.Header{}
	}

	for key := range w.Header() {
		w.Header().Del(key)
	}
	for key, headers := range res.Header {
		if !mime.IsResponseHeader(key) {
			continue
		}
		for _, header := range headers {
			w.Header().Add(key, header)
		}
	}

	buf := bytes.NewBuffer(nil)
	if err := mime.Encode(buf, res.Body, textproto.MIMEHeader(w.Header())); err != nil {
		return err
	}

	status := res.Status
	if status == 0 {
		status = http.StatusOK
	}
	w.WriteHeader(status)

	if _, err := io.Copy(w, buf); err != nil {
		return err
	}

	return nil
}

// NewListenNodeCodec creates a new codec for ListenNodeSpec.
func NewListenNodeCodec() scheme.Codec {
	return scheme.CodecWithType(func(spec *ListenNodeSpec) (node.Node, error) {
		switch spec.Protocol {
		case ProtocolHTTP:
			return NewHTTPListenNode(fmt.Sprintf(":%d", spec.Port)), nil
		}
		return nil, errors.WithStack(ErrInvalidProtocol)
	})
}
