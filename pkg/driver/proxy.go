package driver

import (
	"sync"

	"github.com/pkg/errors"
)

// Proxy wraps a Driver and allows dynamic replacement with concurrency safety.
type Proxy struct {
	driver Driver
	mu     sync.RWMutex
}

var _ Driver = (*Proxy)(nil)

// NewProxy returns a new Proxy that wraps the given driver.
func NewProxy(driver Driver) *Proxy {
	return &Proxy{driver: driver}
}

// Open delegates to the underlying driver's Open method.
func (p *Proxy) Open(name string) (Conn, error) {
	p.mu.RLock()
	defer p.mu.RUnlock()

	if p.driver == nil {
		return nil, errors.WithStack(ErrNotRegistered)
	}
	return p.driver.Open(name)
}

// Wrap sets the underlying driver.
func (p *Proxy) Wrap(driver Driver) {
	p.mu.Lock()
	defer p.mu.Unlock()

	p.driver = driver
}

// Unwrap returns the currently wrapped driver.
func (p *Proxy) Unwrap() Driver {
	p.mu.RLock()
	defer p.mu.RUnlock()

	return p.driver
}

// Close delegates to the underlying driver's Close method if set.
func (p *Proxy) Close() error {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.driver == nil {
		return nil
	}
	return p.driver.Close()
}
