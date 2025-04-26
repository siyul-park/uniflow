package driver

import (
	"sync"
)

// Conn provides access to named Store instances.
type Conn interface {
	Load(name string) (Store, error)
	Close() error
}

type conn struct {
	stores map[string]Store
	mu     sync.Mutex
}

var _ Conn = (*conn)(nil)

func newConn() Conn {
	return &conn{stores: make(map[string]Store)}
}

func (c *conn) Load(name string) (Store, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	s, ok := c.stores[name]
	if !ok {
		s = NewStore()
		c.stores[name] = s
	}
	return s, nil
}

func (c *conn) Close() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.stores = make(map[string]Store)
	return nil
}
