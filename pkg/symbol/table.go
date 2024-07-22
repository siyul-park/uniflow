package symbol

import (
	"slices"
	"sync"

	"github.com/gofrs/uuid"
	"github.com/siyul-park/uniflow/pkg/spec"
)

// TableOptions holds configurations for a Table instance.
type TableOptions struct {
	LoadHooks   []LoadHook   // LoadHooks are functions executed when symbols are loaded.
	UnloadHooks []UnloadHook // UnloadHooks are functions executed when symbols are unloaded.
}

// Table manages symbols, providing storage and operations.
type Table struct {
	symbols     map[uuid.UUID]*Symbol
	namespaces  map[string]map[string]uuid.UUID
	loadHooks   []LoadHook
	unloadHooks []UnloadHook
	mu          sync.RWMutex
}

// NewTable creates a new Table instance.
func NewTable(opts ...TableOptions) *Table {
	var loadHooks []LoadHook
	var unloadHooks []UnloadHook

	for _, opt := range opts {
		loadHooks = append(loadHooks, opt.LoadHooks...)
		unloadHooks = append(unloadHooks, opt.UnloadHooks...)
	}

	return &Table{
		symbols:     make(map[uuid.UUID]*Symbol),
		namespaces:  make(map[string]map[string]uuid.UUID),
		loadHooks:   loadHooks,
		unloadHooks: unloadHooks,
	}
}

// Insert adds a new symbol to the table based on the provided spec.
func (t *Table) Insert(sym *Symbol) error {
	t.mu.Lock()
	defer t.mu.Unlock()

	if sym.refs == nil {
		sym.refs = make(map[string][]spec.Port)
	}

	if _, err := t.free(sym.ID()); err != nil {
		return err
	}
	if err := t.insert(sym); err != nil {
		return err
	}
	return nil
}

// Free removes a symbol from the table by its ID.
func (t *Table) Free(id uuid.UUID) (bool, error) {
	t.mu.Lock()
	defer t.mu.Unlock()

	sym, err := t.free(id)
	if err != nil {
		return false, err
	}
	return sym != nil, nil
}

// Lookup retrieves a symbol from the table by its ID.
func (t *Table) Lookup(id uuid.UUID) (*Symbol, bool) {
	t.mu.RLock()
	defer t.mu.RUnlock()

	sym, ok := t.symbols[id]
	return sym, ok
}

// Keys returns all IDs of symbols in the table.
func (t *Table) Keys() []uuid.UUID {
	t.mu.RLock()
	defer t.mu.RUnlock()

	var ids []uuid.UUID
	for id := range t.symbols {
		ids = append(ids, id)
	}
	return ids
}

// Clear frees all symbols associated with the table.
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
		ns, ok := t.namespaces[sym.Namespace()]
		if !ok {
			ns = make(map[string]uuid.UUID)
			t.namespaces[sym.Namespace()] = ns
		}
		ns[sym.Name()] = sym.ID()
	}

	t.links(sym)
	return t.load(sym)
}

func (t *Table) free(id uuid.UUID) (*Symbol, error) {
	sym, ok := t.symbols[id]
	if !ok {
		return nil, nil
	}

	if err := t.unload(sym); err != nil {
		return nil, err
	}
	t.unlinks(sym)

	if err := sym.Close(); err != nil {
		return nil, err
	}

	if sym.Name() != "" {
		if ns, ok := t.namespaces[sym.Namespace()]; ok {
			delete(ns, sym.Name())
			if len(ns) == 0 {
				delete(t.namespaces, sym.Namespace())
			}
		}
	}
	delete(t.symbols, id)

	return sym, nil
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

func (t *Table) links(sym *Symbol) {
	for name, ports := range sym.Ports() {
		out := sym.Out(name)
		if out == nil {
			continue
		}

		for _, port := range ports {
			id := port.ID
			if id == uuid.Nil {
				id = t.lookup(sym.Namespace(), port.Name)
			}

			if ref, ok := t.symbols[id]; ok {
				if ref.Namespace() == sym.Namespace() {
					if in := ref.In(port.Port); in != nil {
						out.Link(in)
					}

					ref.refs[port.Port] = append(ref.refs[port.Port], spec.Port{
						ID:   sym.ID(),
						Name: port.Name,
						Port: name,
					})
				}
			}
		}
	}

	for _, ref := range t.symbols {
		if ref.Namespace() != sym.Namespace() {
			continue
		}

		for name, ports := range ref.Ports() {
			out := ref.Out(name)
			if out == nil {
				continue
			}

			for _, port := range ports {
				if (port.ID == sym.ID()) || (port.Name != "" && port.Name == sym.Name()) {
					if in := sym.In(port.Port); in != nil {
						out.Link(in)
					}

					sym.refs[port.Port] = append(sym.refs[port.Port], spec.Port{
						ID:   ref.ID(),
						Name: port.Name,
						Port: name,
					})
				}
			}
		}
	}
}

func (t *Table) unlinks(sym *Symbol) {
	for name, ports := range sym.Ports() {
		for _, port := range ports {
			id := port.ID
			if id == uuid.Nil {
				id = t.lookup(sym.Namespace(), port.Name)
			}

			ref, ok := t.symbols[id]
			if !ok {
				continue
			}

			var ports []spec.Port
			for _, port := range ref.refs[port.Port] {
				if port.ID != sym.ID() && port.Port != name {
					ports = append(ports, port)
				}
			}

			if len(ports) > 0 {
				ref.refs[port.Port] = ports
			} else {
				delete(ref.refs, port.Port)
			}
		}
	}
}

func (t *Table) linked(sym *Symbol) []*Symbol {
	nexts := []*Symbol{sym}

	var linked []*Symbol
	for len(nexts) > 0 {
		sym := nexts[len(nexts)-1]
		ok := true
		for _, locations := range sym.refs {
			for _, location := range locations {
				next := t.symbols[location.ID]
				if ok = slices.Contains(nexts, next) || slices.Contains(linked, next); !ok {
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

		for _, ports := range sym.Ports() {
			for _, port := range ports {
				id := port.ID
				if id == uuid.Nil {
					id = t.lookup(sym.Namespace(), port.Name)
				}

				ref, ok := t.symbols[id]
				if !ok || ref.Namespace() != sym.Namespace() {
					return false
				}

				nexts = append(nexts, ref)
			}
		}
	}
	return true
}

func (t *Table) lookup(namespace, name string) uuid.UUID {
	if ns, ok := t.namespaces[namespace]; ok {
		return ns[name]
	}
	return uuid.UUID{}
}
