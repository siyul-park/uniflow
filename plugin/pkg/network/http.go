package network

import (
	"context"
	"net"
	"net/http"
	"net/url"
	"sync"
	"time"

	"github.com/pkg/errors"
	"github.com/siyul-park/uniflow/pkg/node"
	"github.com/siyul-park/uniflow/pkg/port"
	"github.com/siyul-park/uniflow/pkg/primitive"
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
	// TODO: implement
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
