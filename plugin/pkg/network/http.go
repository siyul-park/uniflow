package network

import (
	"context"
	"io"
	"net"
	"net/http"
	"net/url"
	"strconv"
	"sync"
	"time"

	"github.com/pkg/errors"
	"github.com/siyul-park/uniflow/pkg/node"
	"github.com/siyul-park/uniflow/pkg/packet"
	"github.com/siyul-park/uniflow/pkg/port"
	"github.com/siyul-park/uniflow/pkg/primitive"
	"github.com/siyul-park/uniflow/pkg/process"
)

type HTTPNode struct {
	server   *http.Server
	listener net.Listener
	ioPort   *port.Port
	inPort   *port.Port
	outPort  *port.Port
	errPort  *port.Port
	mu       sync.RWMutex
}

type HTTPPayload struct {
	Proto   string          `map:"proto,omitempty"`
	Path    string          `map:"path,omitempty"`
	Method  string          `map:"method,omitempty"`
	Header  http.Header     `map:"header,omitempty"`
	Query   url.Values      `map:"query,omitempty"`
	Cookies []*http.Cookie  `map:"cookies,omitempty"`
	Body    primitive.Value `map:"body,omitempty"`
	Status  int             `map:"status"`
}

type tcpKeepAliveListener struct {
	*net.TCPListener
}

var ErrInvalidListenerNetwork = errors.New("invalid listener network")

var _ node.Node = (*HTTPNode)(nil)
var _ http.Handler = (*HTTPNode)(nil)

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

func (n *HTTPNode) Address() net.Addr {
	n.mu.RLock()
	defer n.mu.RUnlock()
	if n.listener == nil {
		return nil
	}
	return n.listener.Addr()
}

func (n *HTTPNode) WaitForListen(errChan <-chan error) error {
	ctx, cancel := context.WithTimeout(context.Background(), 200*time.Millisecond)
	defer cancel()

	ticker := time.NewTicker(5 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-ticker.C:
			if addr := n.Address(); addr != nil {
				return nil
			}
		case err := <-errChan:
			if err == http.ErrServerClosed {
				return nil
			}
			return err
		}
	}
}

func (n *HTTPNode) Listen() error {
	if err := func() error {
		n.mu.Lock()
		defer n.mu.Unlock()

		if n.listener != nil {
			return nil
		}
		if l, err := newListener(n.server.Addr, "tcp"); err != nil {
			return err
		} else {
			n.listener = l
		}
		return nil
	}(); err != nil {
		return err
	}

	return n.server.Serve(n.listener)
}

func (n *HTTPNode) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	n.mu.RLock()
	defer n.mu.RUnlock()

	proc := process.New()
	defer proc.Exit(nil)

	if err := n.serve(proc, w, r); err != nil {
		// TODO: handle error
		proc.Exit(err)
	}
}

func (n *HTTPNode) Close() error {
	n.mu.Lock()
	defer n.mu.Unlock()

	n.ioPort.Close()
	n.inPort.Close()
	n.outPort.Close()
	n.errPort.Close()

	return nil
}

func (n *HTTPNode) serve(proc *process.Process, w http.ResponseWriter, r *http.Request) error {
	ioStream := n.ioPort.Open(proc)
	inStream := n.inPort.Open(proc)
	outStream := n.outPort.Open(proc)

	if ioStream.Links()+outStream.Links() == 0 {
		return nil
	}

	req, err := n.read(r)
	if err != nil {
		return err
	}

	outPayload, err := primitive.MarshalBinary(req)
	if err != nil {
		return err
	}
	outPck := packet.New(outPayload)

	ioStream.Send(outPck)
	outStream.Send(outPck)

	if ioStream.Links()+inStream.Links() == 0 {
		return nil
	}

	var inPck *packet.Packet
	var ok bool
	select {
	case inPck, ok = <-ioStream.Receive():
	case inPck, ok = <-inStream.Receive():
	}
	if !ok {
		return nil
	}

	inPayload := inPck.Payload()
	var res HTTPPayload
	if err := primitive.Unmarshal(inPayload, &res); err != nil {
		res.Body = inPayload
	}

	return n.write(w, res)
}

func (n *HTTPNode) read(r *http.Request) (HTTPPayload, error) {
	contentType := r.Header.Get(HeaderContentType)

	if b, err := io.ReadAll(r.Body); err != nil {
		return HTTPPayload{}, err
	} else if b, err := UnmarshalMIME(b, &contentType); err != nil {
		return HTTPPayload{}, err
	} else {
		r.Header.Set(HeaderContentType, contentType)
		return HTTPPayload{
			Proto:   r.Proto,
			Path:    r.URL.Path,
			Method:  r.Method,
			Header:  r.Header,
			Query:   r.URL.Query(),
			Cookies: r.Cookies(),
			Body:    b,
		}, nil
	}
}

func (n *HTTPNode) write(w http.ResponseWriter, res HTTPPayload) error {
	contentType := res.Header.Get(HeaderContentType)

	b, err := MarshalMIME(res.Body, &contentType)
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
		if len(b) == 0 {
			status = http.StatusNoContent
		} else {
			status = http.StatusOK
		}
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

func newListener(address, network string) (*tcpKeepAliveListener, error) {
	if network != "tcp" && network != "tcp4" && network != "tcp6" {
		return nil, ErrInvalidListenerNetwork
	}
	l, err := net.Listen(network, address)
	if err != nil {
		return nil, err
	}
	return &tcpKeepAliveListener{l.(*net.TCPListener)}, nil
}

func (ln tcpKeepAliveListener) Accept() (net.Conn, error) {
	if c, err := ln.AcceptTCP(); err != nil {
		return nil, err
	} else if err = c.SetKeepAlive(true); err != nil {
		return nil, err
	} else {
		_ = c.SetKeepAlivePeriod(3 * time.Minute)
		return c, nil
	}
}
