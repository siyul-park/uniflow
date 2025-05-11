package driver

import (
	"sync"
)

// ConnProxy wraps a Conn and allows dynamic replacement with concurrency safety.
type ConnProxy struct {
	conn Conn
	mu   sync.RWMutex
}

var _ Conn = (*ConnProxy)(nil)

// NewConnProxy returns a new ConnProxy that wraps the given Conn.
func NewConnProxy(conn Conn) *ConnProxy {
	return &ConnProxy{conn: conn}
}

// Load delegates to the underlying conn's Load method.
func (p *ConnProxy) Load(name string) (Store, error) {
	p.mu.RLock()
	defer p.mu.RUnlock()

	if p.conn == nil {
		return nil, ErrNotRegistered // or custom error if needed
	}
	return p.conn.Load(name)
}

// Close delegates to the underlying conn's Close method.
func (p *ConnProxy) Close() error {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.conn == nil {
		return nil
	}
	return p.conn.Close()
}

// Wrap sets the underlying Conn.
func (p *ConnProxy) Wrap(conn Conn) {
	p.mu.Lock()
	defer p.mu.Unlock()

	p.conn = conn
}

// Unwrap returns the currently wrapped Conn.
func (p *ConnProxy) Unwrap() Conn {
	p.mu.RLock()
	defer p.mu.RUnlock()

	return p.conn
}
