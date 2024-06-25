package symbol

import (
	"reflect"
	"slices"
	"sync"

	"github.com/gofrs/uuid"
	"github.com/samber/lo"
	"github.com/siyul-park/uniflow/pkg/scheme"
	"github.com/siyul-park/uniflow/pkg/spec"
)

// TableOptions holds options for configuring a Table.
type TableOptions struct {
	LoadHooks   []LoadHook   // LoadHooks define functions to be executed on symbol loading.
	UnloadHooks []UnloadHook // UnloadHooks define functions to be executed on symbol unloading.
}

// Table manages the store and operations for Symbols.
type Table struct {
	scheme      *scheme.Scheme
	symbols     map[uuid.UUID]*Symbol
	namespaces  map[string]map[string]uuid.UUID
	loadHooks   []LoadHook
	unloadHooks []UnloadHook
	mu          sync.RWMutex
}

// NewTable returns a new SymbolTable with the specified options.
func NewTable(sheme *scheme.Scheme, opts ...TableOptions) *Table {
	var loadHooks []LoadHook
	var unloadHooks []UnloadHook

	for _, opt := range opts {
		loadHooks = append(loadHooks, opt.LoadHooks...)
		unloadHooks = append(unloadHooks, opt.UnloadHooks...)
	}

	return &Table{
		scheme:      sheme,
		symbols:     make(map[uuid.UUID]*Symbol),
		namespaces:  make(map[string]map[string]uuid.UUID),
		loadHooks:   loadHooks,
		unloadHooks: unloadHooks,
	}
}

// Insert adds a Symbol to the table.
func (t *Table) Insert(spec spec.Spec) (*Symbol, error) {
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

	if namespace, ok := t.namespaces[namespace]; ok {
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
		t.namespaces[sym.Namespace()] = lo.Assign(t.namespaces[sym.Namespace()], map[string]uuid.UUID{sym.Name(): sym.ID()})
	}

	if err := t.links(sym); err != nil {
		return err
	}

	return nil
}

func (t *Table) free(id uuid.UUID) (*Symbol, error) {
	sym, ok := t.symbols[id]
	if !ok {
		return nil, nil
	}

	if err := t.unlinks(sym); err != nil {
		return nil, err
	}
	if err := sym.Close(); err != nil {
		return nil, err
	}

	if sym.Name() != "" {
		if namespace, ok := t.namespaces[sym.Namespace()]; ok {
			delete(namespace, sym.Name())
			if len(namespace) == 0 {
				delete(t.namespaces, sym.Namespace())
			}
		}
	}
	delete(t.symbols, id)

	return sym, nil
}

func (t *Table) links(sym *Symbol) error {
	for name, locations := range sym.links {
		out := sym.Out(name)
		if out == nil {
			sym.unlinks[name] = locations
			continue
		}

		for _, location := range locations {
			id := location.ID
			if location.Name != "" {
				if namespace, ok := t.namespaces[sym.Namespace()]; ok {
					id = namespace[location.Name]
				}
			}

			if id != (uuid.UUID{}) {
				if ref, ok := t.symbols[id]; ok {
					if ref.Namespace() == sym.Namespace() {
						if in := ref.In(location.Port); in != nil {
							out.Link(in)

							ref.linked[location.Port] = append(ref.linked[location.Port], spec.PortLocation{
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

	for _, ref := range t.symbols {
		if ref.Namespace() != sym.Namespace() {
			continue
		}

		for name, locations := range ref.unlinks {
			out := ref.Out(name)
			if out == nil {
				continue
			}

			for i, location := range locations {
				if (location.ID == sym.ID()) || (location.Name != "" && location.Name == sym.Name()) {
					if in := sym.In(location.Port); in != nil {
						out.Link(in)

						sym.linked[location.Port] = append(sym.linked[location.Port], spec.PortLocation{
							ID:   ref.ID(),
							Name: location.Name,
							Port: name,
						})

						ref.unlinks[name] = append(locations[:i], locations[i+1:]...)
						if len(ref.unlinks[name]) == 0 {
							delete(ref.unlinks, name)
						}
					}
				}
			}
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
				if namespace, ok := t.namespaces[sym.Namespace()]; ok {
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

			var locations []spec.PortLocation
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
			ref := t.symbols[location.ID]

			var unlink spec.PortLocation
			if location.Name == "" {
				unlink = spec.PortLocation{
					ID:   sym.ID(),
					Port: name,
				}
			} else {
				unlink = spec.PortLocation{
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

func (t *Table) load(sym *Symbol) error {
	linked := t.linked(sym)
	for _, sym := range linked {
		if t.active(sym) {
			for _, hook := range t.loadHooks {
				if err := hook.Load(sym); err != nil {
					return err
				}
			}
		}
	}
	return nil
}

func (t *Table) unload(sym *Symbol) error {
	linked := t.linked(sym)
	for i := len(linked) - 1; i >= 0; i-- {
		sym := linked[i]
		if t.active(sym) {
			for j := len(t.unloadHooks) - 1; j >= 0; j-- {
				hook := t.unloadHooks[j]
				if err := hook.Unload(sym); err != nil {
					return err
				}
			}
		}
	}
	return nil
}

func (t *Table) linked(sym *Symbol) []*Symbol {
	nexts := []*Symbol{sym}

	var linked []*Symbol
	for len(nexts) > 0 {
		sym := nexts[len(nexts)-1]
		ok := true
		for _, locations := range sym.linked {
			for _, location := range locations {
				next := t.symbols[location.ID]
				ok = false
				for _, cur := range nexts {
					if ok = ok || cur == next; ok {
						break
					}
				}
				for _, cur := range linked {
					if ok = ok || cur == next; ok {
						break
					}
				}
				if !ok {
					nexts = append(nexts, next)
					break
				}
			}
			if !ok {
				break
			}
		}
		if ok {
			nexts = nexts[0 : len(nexts)-1]
			linked = append(linked, sym)
		}
	}

	slices.Reverse(linked)
	return linked
}

func (t *Table) active(sym *Symbol) bool {
	nexts := []*Symbol{sym}
	visits := map[*Symbol]struct{}{}
	for len(nexts) > 0 {
		sym := nexts[0]
		nexts = nexts[1:]

		if _, visit := visits[sym]; visit {
			continue
		}
		visits[sym] = struct{}{}

		if len(sym.unlinks) > 0 {
			return false
		}

		for _, locations := range sym.links {
			for _, location := range locations {
				id := location.ID
				if location.Name != "" {
					if namespace, ok := t.namespaces[sym.Namespace()]; ok {
						id = namespace[location.Name]
					}
				}

				next := t.symbols[id]
				nexts = append(nexts, next)
			}
		}
	}
	return true
}
