package symbol

import (
	"reflect"
	"sync"

	"github.com/gofrs/uuid"
	"github.com/samber/lo"
	"github.com/siyul-park/uniflow/pkg/scheme"
)

// TableOptions holds options for configuring a Table.

type TableOptions struct {
	LoadHooks   []LoadHook   // LoadHooks define functions to be executed on symbol loading.
	UnloadHooks []UnloadHook // UnloadHooks define functions to be executed on symbol unloading.
}

// Table manages the storage and operations for Symbols.
type Table struct {
	scheme      *scheme.Scheme
	symbols     map[uuid.UUID]*Symbol
	index       map[string]map[string]uuid.UUID
	loadHooks   []LoadHook
	unloadHooks []UnloadHook
	mu          sync.RWMutex
}

// NewTable returns a new SymbolTable with the specified options.
func NewTable(sh *scheme.Scheme, opts ...TableOptions) *Table {
	var loadHooks []LoadHook
	var unloadHooks []UnloadHook

	for _, opt := range opts {
		loadHooks = append(loadHooks, opt.LoadHooks...)
		unloadHooks = append(unloadHooks, opt.UnloadHooks...)
	}

	return &Table{
		scheme:      sh,
		symbols:     make(map[uuid.UUID]*Symbol),
		index:       make(map[string]map[string]uuid.UUID),
		loadHooks:   loadHooks,
		unloadHooks: unloadHooks,
	}
}

// Insert adds a Symbol to the table.
func (t *Table) Insert(spec scheme.Spec) (*Symbol, error) {
	t.mu.Lock()
	defer t.mu.Unlock()

	if sym, ok := t.symbols[spec.GetID()]; ok && reflect.DeepEqual(sym.spec, spec) {
		return nil, nil
	}

	n, err := t.scheme.Decode(spec)
	if err != nil {
		return nil, err
	}

	sym := New(spec, n)

	if _, err := t.free(sym.ID()); err != nil {
		return nil, err
	}
	if err := t.insert(sym); err != nil {
		return nil, err
	}

	return sym, nil
}

// Free removes a Symbol from the table.
func (t *Table) Free(id uuid.UUID) (bool, error) {
	t.mu.Lock()
	defer t.mu.Unlock()

	sym, err := t.free(id)
	if err != nil {
		return false, err
	}
	return sym != nil, nil
}

// LookupByID retrieves a Symbol by its ID.
func (t *Table) LookupByID(id uuid.UUID) (*Symbol, bool) {
	t.mu.RLock()
	defer t.mu.RUnlock()

	sym, ok := t.symbols[id]
	return sym, ok
}

// LookupByName retrieves a Symbol by its namespace and name.
func (t *Table) LookupByName(namespace, name string) (*Symbol, bool) {
	t.mu.RLock()
	defer t.mu.RUnlock()

	if namespace, ok := t.index[namespace]; ok {
		if id, ok := namespace[name]; ok {
			sym, ok := t.symbols[id]
			return sym, ok
		}
	}
	return nil, false
}

// Keys returns all Symbol's ID.
func (t *Table) Keys() []uuid.UUID {
	t.mu.RLock()
	defer t.mu.RUnlock()

	var ids []uuid.UUID
	for id := range t.symbols {
		ids = append(ids, id)
	}

	return ids
}

// Clear free all associated Symbols.
func (t *Table) Clear() error {
	t.mu.Lock()
	defer t.mu.Unlock()

	for id := range t.symbols {
		if _, err := t.free(id); err != nil {
			return err
		}
	}

	return nil
}

func (t *Table) insert(sym *Symbol) error {
	t.symbols[sym.ID()] = sym
	if sym.Name() != "" {
		t.index[sym.Namespace()] = lo.Assign(t.index[sym.Namespace()], map[string]uuid.UUID{sym.Name(): sym.ID()})
	}

	if err := t.links(sym); err != nil {
		return err
	}
	if err := t.relinks(sym); err != nil {
		return err
	}

	return nil
}

func (t *Table) free(id uuid.UUID) (*Symbol, error) {
	sym, ok := t.symbols[id]
	if !ok {
		return nil, nil
	}

	if err := sym.Close(); err != nil {
		return nil, err
	}
	if sym.Name() != "" {
		if namespace, ok := t.index[sym.Namespace()]; ok {
			delete(namespace, sym.Name())
			if len(namespace) == 0 {
				delete(t.index, sym.Namespace())
			}
		}
	}
	delete(t.symbols, id)

	if err := t.unlinks(sym); err != nil {
		return nil, err
	}

	return sym, nil
}

func (t *Table) links(sym *Symbol) error {
	for name, locations := range sym.links {
		p1, ok := sym.Port(name)
		if !ok {
			sym.unlinks[name] = locations
			continue
		}

		for _, location := range locations {
			id := location.ID
			if location.Name != "" {
				if namespace, ok := t.index[sym.Namespace()]; ok {
					id = namespace[location.Name]
				}
			}

			if id != (uuid.UUID{}) {
				if ref, ok := t.symbols[id]; ok {
					if ref.Namespace() == sym.Namespace() {
						if p2, ok := ref.Port(location.Port); ok {
							p1.Link(p2)
							ref.linked[location.Port] = append(ref.linked[location.Port], scheme.PortLocation{
								ID:   sym.ID(),
								Name: location.Name,
								Port: name,
							})
							continue
						}
					}
				}
			}

			sym.unlinks[name] = append(sym.unlinks[name], location)
		}
	}

	return t.load(sym)
}

func (t *Table) unlinks(sym *Symbol) error {
	if err := t.unload(sym); err != nil {
		return err
	}

	for name, locations := range sym.links {
		for _, location := range locations {
			id := location.ID
			if location.Name != "" {
				if namespace, ok := t.index[sym.Namespace()]; ok {
					id = namespace[location.Name]
				}
			}

			if id == (uuid.UUID{}) {
				continue
			}

			ref, ok := t.symbols[id]
			if !ok {
				continue
			}

			var locations []scheme.PortLocation
			for _, location := range ref.linked[location.Port] {
				if location.ID != sym.ID() && location.Port != name {
					locations = append(locations, location)
				}
			}

			if len(locations) > 0 {
				ref.linked[location.Port] = locations
			} else {
				delete(ref.linked, location.Port)
			}
		}
	}

	for name, locations := range sym.linked {
		for i, location := range locations {
			if err := t.unload(t.symbols[location.ID]); err != nil {
				return err
			}

			ref := t.symbols[location.ID]

			var unlink scheme.PortLocation
			if location.Name == "" {
				unlink = scheme.PortLocation{
					ID:   sym.ID(),
					Port: name,
				}
			} else {
				unlink = scheme.PortLocation{
					Name: location.Name,
					Port: name,
				}
			}

			sym.linked[name] = append(locations[:i], locations[i+1:]...)
			ref.unlinks[location.Port] = append(ref.unlinks[location.Port], unlink)
		}
	}

	return nil
}

func (t *Table) relinks(sym *Symbol) error {
	for _, ref := range t.symbols {
		if ref.Namespace() != sym.Namespace() {
			continue
		}

		for name, locations := range ref.unlinks {
			p1, ok := ref.Port(name)
			if !ok {
				continue
			}

			for i, location := range locations {
				if (location.ID == sym.ID()) || (location.Name != "" && location.Name == sym.Name()) {
					if p2, ok := sym.Port(location.Port); ok {
						p1.Link(p2)
						sym.linked[location.Port] = append(sym.linked[location.Port], scheme.PortLocation{
							ID:   ref.ID(),
							Name: location.Name,
							Port: name,
						})
						ref.unlinks[name] = append(locations[:i], locations[i+1:]...)
					}
				}
			}
		}

		if err := t.load(ref); err != nil {
			return err
		}
	}

	return nil
}

func (t *Table) load(sym *Symbol) error {
	if t.shouldSkipLoad(sym) {
		return nil
	}
	for _, hook := range t.loadHooks {
		if err := hook.Load(sym.node); err != nil {
			return err
		}
	}
	return nil
}

func (t *Table) unload(sym *Symbol) error {
	if t.shouldSkipLoad(sym) {
		return nil
	}
	for _, hook := range t.unloadHooks {
		if err := hook.Unload(sym.node); err != nil {
			return err
		}
	}
	return nil
}

func (t *Table) shouldSkipLoad(sym *Symbol) bool {
	return len(sym.unlinks) > 0
}
