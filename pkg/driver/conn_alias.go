package driver

import (
	"sync"
)

// ConnAlias represents a connection with alias support for table names.
type ConnAlias struct {
	conn  Conn
	alias map[string]string
	mu    sync.RWMutex
}

var _ Conn = (*ConnAlias)(nil)

// NewConnAlias creates a new ConnAlias instance with the given connection.
func NewConnAlias(conn Conn) *ConnAlias {
	return &ConnAlias{
		conn:  conn,
		alias: make(map[string]string),
	}
}

// Alias assigns an alias to a table name.
func (c *ConnAlias) Alias(name, alias string) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.alias[alias] = name
}

// Load retrieves a table by its name.
func (c *ConnAlias) Load(name string) (Store, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	origin, ok := c.alias[name]
	if ok {
		name = origin
	}

	return c.conn.Load(name)
}

// Close closes the underlying connection.
func (c *ConnAlias) Close() error {
	return c.conn.Close()
}
