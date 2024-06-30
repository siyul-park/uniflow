package system

import (
	"sync"

	"github.com/pkg/errors"
)

// Table represents a table of system call operations.
type Table struct {
	data map[string]any
	mu   sync.RWMutex
}

var ErrInvalidOperation = errors.New("operation is invalid")

// NewTable creates a new Table instance.
func NewTable() *Table {
	return &Table{
		data: make(map[string]any),
	}
}

// Store adds or updates a system call opcode in the table.
func (t *Table) Store(opcode string, fn any) {
	t.mu.Lock()
	defer t.mu.Unlock()

	t.data[opcode] = fn
}

// Load retrieves a system call opcode from the table.
func (t *Table) Load(opcode string) (any, error) {
	t.mu.RLock()
	defer t.mu.RUnlock()

	fn, ok := t.data[opcode]
	if !ok {
		return nil, ErrInvalidOperation
	}
	return fn, nil
}
