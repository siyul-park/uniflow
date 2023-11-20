package symbol

import (
	"sync"

	"github.com/oklog/ulid/v2"
	"github.com/siyul-park/uniflow/pkg/node"
)

type (
	// TableOptions is a options for Table.
	TableOptions struct {
		PreLoadHooks    []PreLoadHook
		PostLoadHooks   []PostLoadHook
		PreUnloadHooks  []PreUnloadHook
		PostUnloadHooks []PostUnloadHook
	}

	// Table is the storage that manages Symbol.
	Table struct {
		data            map[ulid.ULID]node.Node
		preLoadHooks    []PreLoadHook
		postLoadHooks   []PostLoadHook
		preUnloadHooks  []PreUnloadHook
		postUnloadHooks []PostUnloadHook
		mu              sync.RWMutex
	}
)

// NewTable returns a new SymbolTable
func NewTable(opts ...TableOptions) *Table {
	var preLoadHooks []PreLoadHook
	var postLoadHooks []PostLoadHook
	var preUnloadHooks []PreUnloadHook
	var postUnloadHooks []PostUnloadHook

	for _, opt := range opts {
		preLoadHooks = append(preLoadHooks, opt.PreLoadHooks...)
		postLoadHooks = append(postLoadHooks, opt.PostLoadHooks...)
		preUnloadHooks = append(preUnloadHooks, opt.PreUnloadHooks...)
		postUnloadHooks = append(postUnloadHooks, opt.PostUnloadHooks...)
	}

	return &Table{
		data:            make(map[ulid.ULID]node.Node),
		preLoadHooks:    preLoadHooks,
		postLoadHooks:   postLoadHooks,
		preUnloadHooks:  preUnloadHooks,
		postUnloadHooks: postUnloadHooks,
	}
}

// Insert inserts a node.Node.
func (t *Table) Insert(n node.Node) (node.Node, error) {
	t.mu.Lock()
	defer t.mu.Unlock()

	if origin, ok := t.data[n.ID()]; ok {
		if err := t.preUnload(origin); err != nil {
			return nil, err
		}
		if err := origin.Close(); err != nil {
			return nil, err
		}
		if err := t.postUnload(origin); err != nil {
			return nil, err
		}
	}

	if err := t.preLoad(n); err != nil {
		return nil, err
	}
	t.data[n.ID()] = n
	if err := t.postLoad(n); err != nil {
		return nil, err
	}

	return n, nil
}

// Free removes a Symbol.
func (t *Table) Free(id ulid.ULID) (bool, error) {
	t.mu.Lock()
	defer t.mu.Unlock()

	if n, ok := t.data[id]; ok {
		if err := n.Close(); err != nil {
			return false, err
		}
		delete(t.data, id)
		return true, nil
	}

	return false, nil
}

// Lookup returns a node.Node.
func (t *Table) Lookup(id ulid.ULID) (node.Node, bool) {
	t.mu.RLock()
	defer t.mu.RUnlock()

	n, ok := t.data[id]
	return n, ok
}

// Close closes the SymbolTable.
func (t *Table) Close() error {
	t.mu.Lock()
	defer t.mu.Unlock()

	for id, n := range t.data {
		if err := n.Close(); err != nil {
			return err
		}
		delete(t.data, id)
	}

	return nil
}

func (t *Table) preLoad(n node.Node) error {
	for _, hook := range t.preLoadHooks {
		if err := hook.PreLoad(n); err != nil {
			return err
		}
	}
	return nil
}

func (t *Table) postLoad(n node.Node) error {
	for _, hook := range t.postLoadHooks {
		if err := hook.PostLoad(n); err != nil {
			return err
		}
	}
	return nil
}

func (t *Table) preUnload(n node.Node) error {
	for _, hook := range t.preUnloadHooks {
		if err := hook.PreUnload(n); err != nil {
			return err
		}
	}
	return nil
}

func (t *Table) postUnload(n node.Node) error {
	for _, hook := range t.postUnloadHooks {
		if err := hook.PostUnload(n); err != nil {
			return err
		}
	}
	return nil
}
