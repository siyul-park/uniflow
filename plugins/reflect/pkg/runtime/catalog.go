package runtime

import (
	"github.com/pkg/errors"
	"github.com/siyul-park/sqlbridge/schema"
	"github.com/siyul-park/uniflow/pkg/runtime"
)

// Catalog represents a collection of tables managed by the runtime agent.
type Catalog struct {
	agent *runtime.Agent
}

var _ schema.Catalog = (*Catalog)(nil)

// NewCatalog creates a new Catalog instance with the given agent.
func NewCatalog(agent *runtime.Agent) *Catalog {
	return &Catalog{agent: agent}
}

// Table returns the table corresponding to the given name.
func (c *Catalog) Table(name string) (schema.Table, error) {
	switch name {
	case "frames":
		return NewFrameTable(c.agent), nil
	case "processes":
		return NewProcessTable(c.agent), nil
	case "symbols":
		return NewSymbolTable(c.agent), nil
	default:
		return nil, errors.WithStack(schema.ErrTableNotFound)
	}
}
