package symbol

import (
	"reflect"
	"sync"

	"github.com/gofrs/uuid"
	"github.com/samber/lo"
	"github.com/siyul-park/uniflow/pkg/event"
	"github.com/siyul-park/uniflow/pkg/scheme"
)

// TableOptions holds options for configuring a Table.

type TableOptions struct {
	Broker *event.Broker
}

// Table manages the storage and operations for Symbols.
type Table struct {
	scheme  *scheme.Scheme
	broker  *event.Broker
	symbols map[uuid.UUID]*Symbol
	index   map[string]map[string]uuid.UUID
	mu      sync.RWMutex
}

const TopicLoad = "load"
const TopicUnload = "unload"

// NewTable returns a new SymbolTable with the specified options.
func NewTable(sh *scheme.Scheme, opts ...TableOptions) *Table {
	var broker *event.Broker

	for _, opt := range opts {
		if opt.Broker != nil {
			broker = opt.Broker
		}
	}

	return &Table{
		scheme:  sh,
		broker:  broker,
		symbols: make(map[uuid.UUID]*Symbol),
		index:   make(map[string]map[string]uuid.UUID),
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
	if _, err := t.free(sym.ID()); err != nil {
		return err
	}

	t.symbols[sym.ID()] = sym
	if sym.Name() != "" {
		t.index[sym.Namespace()] = lo.Assign(t.index[sym.Namespace()], map[string]uuid.UUID{sym.Name(): sym.ID()})
	}

	t.links(sym)
	t.relinks(sym)

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

	t.unlinks(sym)

	return sym, nil
}

func (t *Table) links(sym *Symbol) {
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

	t.load(sym)
}

func (t *Table) unlinks(sym *Symbol) {
	t.unload(sym)

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
			t.unload(t.symbols[location.ID])

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
}

func (t *Table) relinks(sym *Symbol) {
	for _, ref := range t.symbols {
		if !t.shouldSkipLoad(ref) || ref.Namespace() != sym.Namespace() {
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
						sym.linked[location.Port] = append(sym.linked[location.Port], scheme.PortLocation{
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

		t.load(ref)
	}
}

func (t *Table) load(sym *Symbol) {
	if t.broker == nil || t.shouldSkipLoad(sym) {
		return
	}

	p := t.broker.Producer(TopicLoad)
	e := event.New(sym.node)

	p.Produce(e)

	<-e.Done()
}

func (t *Table) unload(sym *Symbol) {
	if t.broker == nil || t.shouldSkipLoad(sym) {
		return
	}

	p := t.broker.Producer(TopicUnload)
	e := event.New(sym.node)

	p.Produce(e)

	<-e.Done()
}

func (t *Table) shouldSkipLoad(sym *Symbol) bool {
	return len(sym.unlinks) > 0
}
