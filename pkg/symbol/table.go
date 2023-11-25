package symbol

import (
	"sync"

	"github.com/oklog/ulid/v2"
	"github.com/samber/lo"
	"github.com/siyul-park/uniflow/pkg/node"
	"github.com/siyul-park/uniflow/pkg/scheme"
)

type (
	// TableOptions is a options for Table.
	TableOptions struct {
		LoadHooks   []LoadHook
		UnloadHooks []UnloadHook
	}

	// Table is the storage that manages Symbol.
	Table struct {
		nodes       map[ulid.ULID]node.Node
		specs       map[ulid.ULID]scheme.Spec
		unlinks     map[ulid.ULID]map[string][]scheme.PortLocation
		linked      map[ulid.ULID]map[string][]scheme.PortLocation
		index       map[string]map[string]ulid.ULID
		loadHooks   []LoadHook
		unloadHooks []UnloadHook
		mu          sync.RWMutex
	}
)

// NewTable returns a new SymbolTable
func NewTable(opts ...TableOptions) *Table {
	var loadHooks []LoadHook
	var unloadHooks []UnloadHook

	for _, opt := range opts {
		loadHooks = append(loadHooks, opt.LoadHooks...)
		unloadHooks = append(unloadHooks, opt.UnloadHooks...)
	}

	return &Table{
		nodes:       make(map[ulid.ULID]node.Node),
		specs:       make(map[ulid.ULID]scheme.Spec),
		unlinks:     make(map[ulid.ULID]map[string][]scheme.PortLocation),
		linked:      make(map[ulid.ULID]map[string][]scheme.PortLocation),
		index:       make(map[string]map[string]ulid.ULID),
		loadHooks:   loadHooks,
		unloadHooks: unloadHooks,
	}
}

// Insert inserts a node.Node.
func (t *Table) Insert(n node.Node, spec scheme.Spec) error {
	t.mu.Lock()
	defer t.mu.Unlock()

	prevNode := t.nodes[n.ID()]
	prevSpec := t.specs[n.ID()]

	if prevNode != nil {
		if len(t.unlinks[n.ID()]) == 0 {
			if err := t.unload(prevNode); err != nil {
				return err
			}
		}
		if n != prevNode {
			if err := prevNode.Close(); err != nil {
				return err
			}
		}
	}

	t.nodes[n.ID()] = n
	t.specs[n.ID()] = spec
	if prevSpec != nil && prevSpec.GetName() != "" {
		if namespace, ok := t.index[prevSpec.GetNamespace()]; ok {
			delete(namespace, prevSpec.GetName())
			if len(namespace) == 0 {
				delete(t.index, prevSpec.GetNamespace())
			}
		}
	}
	t.index[spec.GetNamespace()] = lo.Assign(t.index[spec.GetNamespace()], map[string]ulid.ULID{spec.GetName(): n.ID()})

	var deletions map[string][]scheme.PortLocation
	if prevSpec != nil {
		deletions = prevSpec.GetLinks()
	}
	additions := spec.GetLinks()
	unlinks := map[string][]scheme.PortLocation{}

	for name, locations := range deletions {
		for _, location := range locations {
			id := location.ID
			if id == (ulid.ULID{}) {
				if location.Name != "" {
					if namespace, ok := t.index[prevSpec.GetNamespace()]; ok {
						id = namespace[location.Name]
					}
				}
			}

			if id == (ulid.ULID{}) {
				continue
			}

			linked := t.linked[id]
			var locations []scheme.PortLocation
			for _, location := range linked[location.Port] {
				if location.ID != n.ID() && location.Port != name {
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

	for name, locations := range additions {
		p1, ok := n.Port(name)
		if !ok {
			unlinks[name] = locations
			continue
		}

		for _, location := range locations {
			id := location.ID
			if id == (ulid.ULID{}) {
				if location.Name != "" {
					if namespace, ok := t.index[spec.GetNamespace()]; ok {
						id = namespace[location.Name]
					}
				}
			}

			if id == (ulid.ULID{}) {
				unlinks[name] = append(unlinks[name], location)
				continue
			}

			if ref, ok := t.specs[id]; ok {
				if ref.GetNamespace() != spec.GetNamespace() {
					continue
				}
			}

			if ref, ok := t.nodes[id]; ok {
				if p2, ok := ref.Port(location.Port); ok {
					p1.Link(p2)

					linked := t.linked[ref.ID()]
					if linked == nil {
						linked = make(map[string][]scheme.PortLocation)
					}
					linked[location.Port] = append(linked[location.Port], scheme.PortLocation{
						ID:   n.ID(),
						Port: name,
					})
					t.linked[ref.ID()] = linked

					continue
				}
			}
			unlinks[name] = append(unlinks[name], location)
		}
	}

	if len(unlinks) > 0 {
		t.unlinks[n.ID()] = unlinks
	} else {
		delete(t.unlinks, n.ID())
		if err := t.load(n); err != nil {
			return err
		}
	}

	for name, locations := range t.linked[n.ID()] {
		p1, ok := n.Port(name)
		if !ok {
			continue
		}
		for _, location := range locations {
			ref := t.nodes[location.ID]
			if p2, ok := ref.Port(location.Port); ok {
				p1.Link(p2)
			}
		}
	}

	for id, unlinks := range t.unlinks {
		if ref := t.specs[id]; ref.GetNamespace() != spec.GetNamespace() {
			continue
		}

		ref := t.nodes[id]
		for name, locations := range unlinks {
			p1, ok := ref.Port(name)
			if !ok {
				continue
			}

			for i, location := range locations {
				if (location.ID == spec.GetID()) || (location.Name != "" && location.Name == spec.GetName()) {
					if p2, ok := n.Port(location.Port); ok {
						p1.Link(p2)

						linked := t.linked[n.ID()]
						if linked == nil {
							linked = make(map[string][]scheme.PortLocation)
						}
						linked[location.Port] = append(linked[location.Port], scheme.PortLocation{
							ID:   ref.ID(),
							Port: name,
						})
						t.linked[n.ID()] = linked

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
			if err := t.load(n); err != nil {
				return err
			}
		}
	}

	return nil
}

// Free removes a Symbol.
func (t *Table) Free(id ulid.ULID) (bool, error) {
	t.mu.Lock()
	defer t.mu.Unlock()

	if n, ok := t.nodes[id]; ok {
		if err := n.Close(); err != nil {
			return false, err
		}

		spec := t.specs[id]
		if namespace, ok := t.index[spec.GetNamespace()]; ok {
			delete(namespace, spec.GetName())
			if len(namespace) == 0 {
				delete(t.index, spec.GetNamespace())
			}
		}

		for name, locations := range t.linked[id] {
			for _, location := range locations {
				unlinks := t.unlinks[location.ID]
				if unlinks == nil {
					unlinks = make(map[string][]scheme.PortLocation)
				}
				unlinks[location.Port] = append(unlinks[location.Port], scheme.PortLocation{
					ID:   id,
					Port: name,
				})
				t.unlinks[location.ID] = unlinks
			}
		}

		delete(t.nodes, id)
		delete(t.specs, id)
		delete(t.unlinks, id)
		delete(t.linked, id)

		t.unload(n)

		return true, nil
	}

	return false, nil
}

// Lookup returns a node.Node.
func (t *Table) Lookup(id ulid.ULID) (node.Node, bool) {
	t.mu.RLock()
	defer t.mu.RUnlock()

	n, ok := t.nodes[id]
	return n, ok
}

// Close closes the SymbolTable.
func (t *Table) Close() error {
	t.mu.Lock()
	defer t.mu.Unlock()

	for id, n := range t.nodes {
		if err := n.Close(); err != nil {
			return err
		}
		delete(t.nodes, id)
	}
	t.specs = make(map[ulid.ULID]scheme.Spec)
	t.unlinks = make(map[ulid.ULID]map[string][]scheme.PortLocation)
	t.linked = make(map[ulid.ULID]map[string][]scheme.PortLocation)
	t.index = make(map[string]map[string]ulid.ULID)

	return nil
}

func (t *Table) load(n node.Node) error {
	for _, hook := range t.loadHooks {
		if err := hook.Load(n); err != nil {
			return err
		}
	}
	return nil
}

func (t *Table) unload(n node.Node) error {
	for _, hook := range t.unloadHooks {
		if err := hook.Unload(n); err != nil {
			return err
		}
	}
	return nil
}
