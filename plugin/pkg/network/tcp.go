package network

import (
	"net"
	"time"

	"github.com/pkg/errors"
)

type tcpKeepAliveListener struct {
	*net.TCPListener
}

var ErrInvalidListenerNetwork = errors.New("invalid listener network")

func newTCPKeepAliveListener(address, network string) (*tcpKeepAliveListener, error) {
	if network != "tcp" && network != "tcp4" && network != "tcp6" {
		return nil, ErrInvalidListenerNetwork
	}
	l, err := net.Listen(network, address)
	if err != nil {
		return nil, err
	}
	return &tcpKeepAliveListener{l.(*net.TCPListener)}, nil
}

func (l tcpKeepAliveListener) Accept() (net.Conn, error) {
	if c, err := l.AcceptTCP(); err != nil {
		return nil, err
	} else if err = c.SetKeepAlive(true); err != nil {
		return nil, err
	} else {
		_ = c.SetKeepAlivePeriod(3 * time.Minute)
		return c, nil
	}
}
