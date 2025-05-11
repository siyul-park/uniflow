package driver

import (
	"github.com/pkg/errors"
	"github.com/siyul-park/sqlbridge/schema"
	"github.com/siyul-park/uniflow/pkg/driver"
	"github.com/siyul-park/uniflow/pkg/spec"
	"github.com/siyul-park/uniflow/pkg/value"
)

// Catalog manages a collection of tables.
type Catalog struct {
	conn driver.Conn
}

var _ schema.Catalog = (*Catalog)(nil)

// NewCatalog creates a new Catalog with the given connection.
func NewCatalog(conn driver.Conn) *Catalog {
	return &Catalog{conn: conn}
}

// Table retrieves the table for the given name.
func (c *Catalog) Table(name string) (schema.Table, error) {
	store, err := c.conn.Load(name)
	if err != nil {
		return nil, err
	}
	switch name {
	case "specs":
		return NewTable[spec.Spec](store), nil
	case "values":
		return NewTable[*value.Value](store), nil
	default:
		return nil, errors.WithStack(schema.ErrTableNotFound)
	}
}
