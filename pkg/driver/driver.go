package driver

import "sync"

// Driver manages the opening and closing of connections.
type Driver interface {
	Open(name string) (Conn, error)
	Close() error
}

type driver struct {
	conns map[string]Conn
	mu    sync.Mutex
}

var _ Driver = (*driver)(nil)

// New creates a new driver instance with an empty connections map.
func New() Driver {
	return &driver{conns: make(map[string]Conn)}
}

func (d *driver) Open(name string) (Conn, error) {
	d.mu.Lock()
	defer d.mu.Unlock()

	c, ok := d.conns[name]
	if !ok {
		c = newConn()
		d.conns[name] = c
	}
	return c, nil
}

func (d *driver) Close() error {
	for _, c := range d.conns {
		if err := c.Close(); err != nil {
			return err
		}
	}
	d.conns = make(map[string]Conn)
	return nil
}
