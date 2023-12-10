package symbol

import (
	"reflect"
	"sync"

	"github.com/oklog/ulid/v2"
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
	symbols     map[ulid.ULID]*Symbol
	unlinks     map[ulid.ULID]map[string][]scheme.PortLocation
	linked      map[ulid.ULID]map[string][]scheme.PortLocation
	index       map[string]map[string]ulid.ULID
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
		symbols:     make(map[ulid.ULID]*Symbol),
		unlinks:     make(map[ulid.ULID]map[string][]scheme.PortLocation),
		linked:      make(map[ulid.ULID]map[string][]scheme.PortLocation),
		index:       make(map[string]map[string]ulid.ULID),
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

	sym := &Symbol{spec: spec, node: n}

	if _, err := t.free(sym.ID()); err != nil {
		return nil, err
	}
	if err := t.insert(sym); err != nil {
		return nil, err
	}

	return sym, nil
}

// Free removes a Symbol from the table.
func (t *Table) Free(id ulid.ULID) (bool, error) {
	t.mu.Lock()
	defer t.mu.Unlock()

	sym, err := t.free(id)
	if err != nil {
		return false, err
	}
	return sym != nil, nil
}

// LookupByID retrieves a Symbol by its ID.
func (t *Table) LookupByID(id ulid.ULID) (*Symbol, bool) {
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

// Close closes the SymbolTable, closing all associated symbols.
func (t *Table) Close() error {
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
		t.index[sym.Namespace()] = lo.Assign(t.index[sym.Namespace()], map[string]ulid.ULID{sym.Name(): sym.ID()})
	}

	unlinks := map[string][]scheme.PortLocation{}

	for name, locations := range sym.spec.GetLinks() {
		p1, ok := sym.Port(name)
		if !ok {
			unlinks[name] = locations
			continue
		}

		for _, location := range locations {
			id := location.ID
			if location.Name != "" {
				if namespace, ok := t.index[sym.Namespace()]; ok {
					id = namespace[location.Name]
				}
			}

			if id != (ulid.ULID{}) {
				if ref, ok := t.symbols[id]; ok {
					if ref.Namespace() == sym.Namespace() {
						if p2, ok := ref.Port(location.Port); ok {
							p1.Link(p2)

							linked := t.linked[ref.ID()]
							if linked == nil {
								linked = make(map[string][]scheme.PortLocation)
							}
							linked[location.Port] = append(linked[location.Port], scheme.PortLocation{
								ID:   sym.ID(),
								Name: location.Name,
								Port: name,
							})
							t.linked[ref.ID()] = linked

							continue
						}
					}
				}
			}

			unlinks[name] = append(unlinks[name], location)
		}
	}

	if len(unlinks) > 0 {
		t.unlinks[sym.ID()] = unlinks
	}

	if err := t.load(sym); err != nil {
		return err
	}

	for id, unlinks := range t.unlinks {
		ref := t.symbols[id]

		if ref.Namespace() != sym.Namespace() {
			continue
		}

		for name, locations := range unlinks {
			p1, ok := ref.Port(name)
			if !ok {
				continue
			}

			for i, location := range locations {
				if (location.ID == sym.ID()) || (location.Name != "" && location.Name == sym.Name()) {
					if p2, ok := sym.Port(location.Port); ok {
						p1.Link(p2)

						linked := t.linked[sym.ID()]
						if linked == nil {
							linked = make(map[string][]scheme.PortLocation)
						}
						linked[location.Port] = append(linked[location.Port], scheme.PortLocation{
							ID:   ref.ID(),
							Name: location.Name,
							Port: name,
						})
						t.linked[sym.ID()] = linked

						unlinks[name] = append(locations[:i], locations[i+1:]...)
					}
				}
			}

			if len(unlinks[name]) == 0 {
				delete(unlinks, name)
			}
		}

		if len(unlinks) > 0 {
			t.unlinks[id] = unlinks
		} else {
			delete(t.unlinks, id)
		}

		if err := t.load(ref); err != nil {
			return err
		}
	}

	return nil
}

func (t *Table) free(id ulid.ULID) (*Symbol, error) {
	sym, ok := t.symbols[id]
	if !ok {
		return nil, nil
	}

	if err := t.unload(sym); err != nil {
		return nil, err
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

	for name, locations := range sym.spec.GetLinks() {
		for _, location := range locations {
			id := location.ID
			if location.Name != "" {
				if namespace, ok := t.index[sym.Namespace()]; ok {
					id = namespace[location.Name]
				}
			}

			if id == (ulid.ULID{}) {
				continue
			}

			linked := t.linked[id]
			var locations []scheme.PortLocation
			for _, location := range linked[location.Port] {
				if location.ID != sym.ID() && location.Port != name {
					locations = append(locations, location)
				}
			}
			if len(locations) > 0 {
				linked[location.Port] = locations
				t.linked[id] = linked
			} else if len(linked) > 0 {
				delete(linked, location.Port)
				t.linked[id] = linked
			}
		}
	}

	for name, locations := range t.linked[id] {
		for _, location := range locations {
			if err := t.unload(t.symbols[location.ID]); err != nil {
				return nil, err
			}

			unlinks := t.unlinks[location.ID]
			if unlinks == nil {
				unlinks = make(map[string][]scheme.PortLocation)
			}

			var unlink scheme.PortLocation
			if location.Name == "" {
				unlink = scheme.PortLocation{
					ID:   id,
					Port: name,
				}
			} else {
				unlink = scheme.PortLocation{
					Name: location.Name,
					Port: name,
				}
			}

			unlinks[location.Port] = append(unlinks[location.Port], unlink)
			t.unlinks[location.ID] = unlinks
		}
	}

	delete(t.symbols, id)
	delete(t.unlinks, id)
	delete(t.linked, id)

	return sym, nil
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
	return len(t.unlinks[sym.ID()]) > 0
}
