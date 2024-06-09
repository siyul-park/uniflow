package symbol

import (
	"reflect"
	"sync"

	"github.com/gofrs/uuid"
	"github.com/samber/lo"
	"github.com/siyul-park/uniflow/pkg/spec"
)

// TableOptions holds options for configuring a Table.

type TableOptions struct {
	LoadHooks   []LoadHook   // LoadHooks define functions to be executed on symbol loading.
	UnloadHooks []UnloadHook // UnloadHooks define functions to be executed on symbol unloading.
}

// Table manages the storage and operations for Symbols.
type Table struct {
	scheme      *spec.Scheme
	symbols     map[uuid.UUID]*Symbol
	index       map[string]map[string]uuid.UUID
	loadHooks   []LoadHook
	unloadHooks []UnloadHook
	mu          sync.RWMutex
}

// NewTable returns a new SymbolTable with the specified options.
func NewTable(sh *spec.Scheme, opts ...TableOptions) *Table {
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
		if namespace, ok := t.index[sym.Namespace()]; ok {
			delete(namespace, sym.Name())
			if len(namespace) == 0 {
				delete(t.index, sym.Namespace())
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
				if namespace, ok := t.index[sym.Namespace()]; ok {
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

	if err := t.load(sym); err != nil {
		return err
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

		if err := t.load(ref); err != nil {
			return err
		}
	}

	return nil
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
	next := []uuid.UUID{sym.ID()}
	visits := map[uuid.UUID]struct{}{}
	for len(next) > 0 {
		id := next[0]
		next = next[1:]

		if _, ok := visits[id]; ok {
			continue
		}
		visits[id] = struct{}{}

		sym, ok := t.symbols[id]
		if !ok {
			continue
		}
		if sym.Status() == StatusReady {
			continue
		}

		for _, locations := range sym.linked {
			for _, location := range locations {
				next = append(next, location.ID)
			}
		}

		if t.shouldSkipLoad(sym) {
			continue
		}

		for _, hook := range t.loadHooks {
			if err := hook.Load(sym); err != nil {
				return err
			}
		}

		sym.status = StatusReady
	}
	return nil
}

func (t *Table) unload(sym *Symbol) error {
	next := []uuid.UUID{sym.ID()}
	visits := map[uuid.UUID]struct{}{}
	for len(next) > 0 {
		id := next[0]
		next = next[1:]

		if _, ok := visits[id]; ok {
			continue
		}
		visits[id] = struct{}{}

		sym, ok := t.symbols[id]
		if !ok {
			continue
		}
		if sym.Status() == StatusNotReady {
			continue
		}

		for _, locations := range sym.linked {
			for _, location := range locations {
				next = append(next, location.ID)
			}
		}

		for _, hook := range t.unloadHooks {
			if err := hook.Unload(sym); err != nil {
				return err
			}
		}

		sym.status = StatusNotReady
	}
	return nil
}

func (t *Table) shouldSkipLoad(sym *Symbol) bool {
	next := []uuid.UUID{sym.ID()}
	visits := map[uuid.UUID]struct{}{}
	for len(next) > 0 {
		id := next[0]
		next = next[1:]

		if _, ok := visits[id]; ok {
			continue
		}
		visits[id] = struct{}{}

		sym, ok := t.symbols[id]
		if !ok {
			continue
		}

		if len(sym.unlinks) > 0 {
			return true
		}

		for _, locations := range sym.links {
			for _, location := range locations {
				id := location.ID
				if location.Name != "" {
					if namespace, ok := t.index[sym.Namespace()]; ok {
						id = namespace[location.Name]
					}
				}
				next = append(next, id)
			}
		}
	}
	return false
}
