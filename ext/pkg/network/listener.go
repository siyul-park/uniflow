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
	"github.com/siyul-park/uniflow/pkg/packet"
	"github.com/siyul-park/uniflow/pkg/port"
	"github.com/siyul-park/uniflow/pkg/process"
	"github.com/siyul-park/uniflow/pkg/scheme"
	"github.com/siyul-park/uniflow/pkg/spec"
	"github.com/siyul-park/uniflow/pkg/types"
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
	Host      string `map:"host,omitempty"`
	Port      int    `map:"port"`
}

const KindListener = "listener"

const (
	KeyHTTPRequest        = "http.Request"
	KeyHTTPResponseWriter = "http.ResponseWriter"
)

var _ node.Node = (*HTTPListenNode)(nil)
var _ http.Handler = (*HTTPListenNode)(nil)

// NewHTTPListenNode creates a new HTTPListenNode with the specified address.
func NewHTTPListenNode(address string) *HTTPListenNode {
	n := &HTTPListenNode{
		outPort: port.NewOut(),
		errPort: port.NewOut(),
	}
	n.server = &http.Server{
		Addr:    address,
		Handler: n,
	}
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
	n.mu.Lock()
	defer n.mu.Unlock()

	if n.listener != nil {
		return nil
	}

	listener, err := net.Listen("tcp", n.server.Addr)
	if err != nil {
		return err
	}
	n.listener = listener

	go n.server.Serve(n.listener)

	return nil
}

// Shutdown shuts down the HTTPListenNode by closing the server and its associated listener.
func (n *HTTPListenNode) Shutdown() error {
	n.mu.Lock()
	defer n.mu.Unlock()

	if n.listener == nil {
		return nil
	}

	err := n.server.Close()
	n.listener.Close()

	n.server = &http.Server{
		Addr:    n.server.Addr,
		Handler: n.server.Handler,
	}
	n.listener = nil

	return err
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

	var outPck, errPck *packet.Packet
	req, err := n.read(r)
	if err != nil {
		errPck = packet.New(types.NewError(err))
	} else if outPayload, err := types.TextEncoder.Encode(req); err != nil {
		errPck = packet.New(types.NewError(err))
	} else {
		outPck = packet.New(outPayload)
	}

	var backPck *packet.Packet
	if errPck != nil {
		backPck = packet.Call(errWriter, errPck)
	} else {
		backPck = packet.Call(outWriter, outPck)
		if _, ok := backPck.Payload().(types.Error); ok {
			backPck = packet.CallOrFallback(errWriter, backPck, backPck)
		}
	}

	if backPck != packet.None {
		var res *HTTPPayload
		if _, ok := backPck.Payload().(types.Error); ok {
			res = NewHTTPPayload(http.StatusInternalServerError)
		} else if err := types.Decoder.Decode(backPck.Payload(), &res); err != nil {
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
		accept := req.Header.Get(mime.HeaderAcceptEncoding)
		encoding := mime.Negotiate(accept, []string{mime.EncodingIdentity, mime.EncodingGzip, mime.EncodingDeflate, mime.EncodingBr})
		if encoding != "" {
			res.Header.Set(mime.HeaderContentEncoding, encoding)
		}
	}

	if res.Header.Get(mime.HeaderContentType) == "" {
		accept := req.Header.Get(mime.HeaderAccept)
		offers := mime.DetectTypes(res.Body)
		contentType := mime.Negotiate(accept, offers)
		if contentType != "" {
			res.Header.Set(mime.HeaderContentType, contentType)
		}
	}
}

func (n *HTTPListenNode) read(r *http.Request) (*HTTPPayload, error) {
	body, err := mime.Decode(r.Body, textproto.MIMEHeader(r.Header))
	if err != nil {
		return nil, err
	}
	return &HTTPPayload{
		Method:   r.Method,
		Scheme:   r.URL.Scheme,
		Host:     r.Host,
		Path:     r.URL.Path,
		Query:    r.URL.Query(),
		Protocol: r.Proto,
		Header:   r.Header,
		Body:     body,
	}, nil
}

func (n *HTTPListenNode) write(w http.ResponseWriter, res *HTTPPayload) error {
	if res == nil {
		return nil
	}

	h := w.Header()
	for key := range h {
		h.Del(key)
	}
	for key, headers := range res.Header {
		if !mime.IsResponseHeader(key) {
			continue
		}
		for _, header := range headers {
			h.Add(key, header)
		}
	}

	buf := bytes.NewBuffer(nil)
	if err := mime.Encode(buf, res.Body, textproto.MIMEHeader(h)); err != nil {
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
			return NewHTTPListenNode(fmt.Sprintf("%s:%d", spec.Host, spec.Port)), nil
		}
		return nil, errors.WithStack(ErrInvalidProtocol)
	})
}
