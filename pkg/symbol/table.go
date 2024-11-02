package symbol

import (
	"slices"
	"sync"

	"github.com/gofrs/uuid"
	"github.com/siyul-park/uniflow/pkg/node"
	"github.com/siyul-park/uniflow/pkg/port"
	"github.com/siyul-park/uniflow/pkg/spec"
	"github.com/siyul-park/uniflow/pkg/types"
)

// TableOption holds configurations for a Table instance.
type TableOption struct {
	LoadHooks   []LoadHook   // LoadHooks are functions executed when symbols are loaded.
	UnloadHooks []UnloadHook // UnloadHooks are functions executed when symbols are unloaded.
}

// Table manages symbols, providing storage and operations.
type Table struct {
	symbols     map[uuid.UUID]*Symbol
	namespaces  map[string]map[string]uuid.UUID
	references  map[uuid.UUID]map[string][]spec.Port
	loadHooks   LoadHooks
	unloadHooks UnloadHooks
	mu          sync.RWMutex
}

// NewTable creates a new Table instance.
func NewTable(opts ...TableOption) *Table {
	var loadHooks []LoadHook
	var unloadHooks []UnloadHook
	for _, opt := range opts {
		loadHooks = append(loadHooks, opt.LoadHooks...)
		unloadHooks = append(unloadHooks, opt.UnloadHooks...)
	}

	return &Table{
		symbols:     make(map[uuid.UUID]*Symbol),
		namespaces:  make(map[string]map[string]uuid.UUID),
		references:  make(map[uuid.UUID]map[string][]spec.Port),
		loadHooks:   loadHooks,
		unloadHooks: unloadHooks,
	}
}

// Insert adds a new symbol to the table based on the provided spec.
func (t *Table) Insert(sb *Symbol) error {
	t.mu.Lock()
	defer t.mu.Unlock()

	if _, err := t.free(sb.ID()); err != nil {
		return err
	}
	return t.insert(sb)
}

// Free removes a symbol from the table by its ID.
func (t *Table) Free(id uuid.UUID) (bool, error) {
	t.mu.Lock()
	defer t.mu.Unlock()

	sb, err := t.free(id)
	if err != nil {
		return false, err
	}
	return sb != nil, nil
}

// Lookup retrieves a symbol from the table by its ID.
func (t *Table) Lookup(id uuid.UUID) *Symbol {
	t.mu.RLock()
	defer t.mu.RUnlock()

	return t.symbols[id]
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

// Close frees all symbols associated with the table.
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

func (t *Table) insert(sb *Symbol) error {
	t.symbols[sb.ID()] = sb

	if sb.Name() != "" {
		ns, ok := t.namespaces[sb.Namespace()]
		if !ok {
			ns = make(map[string]uuid.UUID)
			t.namespaces[sb.Namespace()] = ns
		}
		ns[sb.Name()] = sb.ID()
	}

	t.links(sb)
	return t.load(sb)
}

func (t *Table) free(id uuid.UUID) (*Symbol, error) {
	sb, ok := t.symbols[id]
	if !ok {
		return nil, nil
	}

	if err := t.unload(sb); err != nil {
		return nil, err
	}
	t.unlinks(sb)

	if err := sb.Close(); err != nil {
		return nil, err
	}

	if sb.Name() != "" {
		if ns, ok := t.namespaces[sb.Namespace()]; ok {
			delete(ns, sb.Name())
			if len(ns) == 0 {
				delete(t.namespaces, sb.Namespace())
			}
		}
	}

	delete(t.symbols, id)

	return sb, nil
}

func (t *Table) load(sb *Symbol) error {
	linked := t.linked(sb)
	for _, sb := range linked {
		if t.active(sb) {
			if err := t.call(sb, node.PortInit); err != nil {
				return err
			}
			if err := t.loadHooks.Load(sb); err != nil {
				return err
			}
		}
	}
	return nil
}

func (t *Table) unload(sb *Symbol) error {
	linked := t.linked(sb)
	for i := len(linked) - 1; i >= 0; i-- {
		sb := linked[i]
		if t.active(sb) {
			if err := t.unloadHooks.Unload(sb); err != nil {
				return err
			}
			if err := t.call(sb, node.PortTerm); err != nil {
				return err
			}
		}
	}
	return nil
}

func (t *Table) links(sb *Symbol) {
	for name, ports := range sb.Ports() {
		out := sb.Out(name)

		for _, port := range ports {
			id := port.ID
			if id == uuid.Nil {
				id = t.lookup(sb.Namespace(), port.Name)
			}

			if ref, ok := t.symbols[id]; ok {
				if ref.Namespace() == sb.Namespace() {
					in := ref.In(port.Port)
					if out != nil && in != nil {
						out.Link(in)
					}

					refences := t.references[ref.ID()]
					if refences == nil {
						refences = make(map[string][]spec.Port)
						t.references[ref.ID()] = refences
					}

					refences[port.Port] = append(refences[port.Port], spec.Port{
						ID:   sb.ID(),
						Name: port.Name,
						Port: name,
					})
				}
			}
		}
	}

	for _, ref := range t.symbols {
		if ref.Namespace() != sb.Namespace() {
			continue
		}

		for name, ports := range ref.Ports() {
			out := ref.Out(name)

			for _, port := range ports {
				if (port.ID == sb.ID()) || (port.Name != "" && port.Name == sb.Name()) {
					in := sb.In(port.Port)
					if out != nil && in != nil {
						out.Link(in)
					}

					refences := t.references[sb.ID()]
					if refences == nil {
						refences = make(map[string][]spec.Port)
						t.references[sb.ID()] = refences
					}

					refences[port.Port] = append(refences[port.Port], spec.Port{
						ID:   ref.ID(),
						Name: port.Name,
						Port: name,
					})
				}
			}
		}
	}
}

func (t *Table) unlinks(sb *Symbol) {
	for name, ports := range sb.Ports() {
		for _, port := range ports {
			id := port.ID
			if id == uuid.Nil {
				id = t.lookup(sb.Namespace(), port.Name)
			}

			ref, ok := t.symbols[id]
			if !ok {
				continue
			}

			references := t.references[ref.ID()]
			if references == nil {
				references = make(map[string][]spec.Port)
				t.references[ref.ID()] = references
			}

			var ports []spec.Port
			for _, port := range references[port.Port] {
				if port.ID != sb.ID() && port.Port != name {
					ports = append(ports, port)
				}
			}

			if len(ports) > 0 {
				references[port.Port] = ports
			} else {
				delete(references, port.Port)
			}
		}
	}

	delete(t.references, sb.ID())
}

func (t *Table) linked(sb *Symbol) []*Symbol {
	degree := map[*Symbol]int{}
	visited := map[*Symbol]struct{}{}
	queue := []*Symbol{sb}
	for len(queue) > 0 {
		curr := queue[0]
		queue = queue[1:]

		if _, ok := visited[curr]; ok {
			continue
		}
		visited[curr] = struct{}{}

		for _, ports := range t.references[curr.ID()] {
			for _, port := range ports {
				id := port.ID
				if id == uuid.Nil {
					id = t.lookup(curr.Namespace(), port.Name)
				}

				if next, ok := t.symbols[id]; ok {
					degree[next]++
					queue = append(queue, next)
				}
			}
		}
	}

	var linked []*Symbol
	queue = []*Symbol{sb}
	for len(queue) > 0 {
		curr := queue[0]
		queue = queue[1:]

		if slices.Contains(linked, curr) {
			continue
		}
		linked = append(linked, curr)

		for _, ports := range t.references[curr.ID()] {
			for _, port := range ports {
				id := port.ID
				if id == uuid.Nil {
					id = t.lookup(curr.Namespace(), port.Name)
				}

				if next, ok := t.symbols[id]; ok {
					degree[next]--
					if degree[next] == 0 {
						queue = append(queue, next)
					}
				}
			}
		}
	}
	for curr, d := range degree {
		if d != 0 {
			linked = append(linked, curr)
		}
	}

	return linked
}

func (t *Table) active(sb *Symbol) bool {
	stack := []*Symbol{sb}
	visited := map[*Symbol]struct{}{}
	for len(stack) > 0 {
		curr := stack[len(stack)-1]
		stack = stack[:len(stack)-1]

		if _, ok := visited[curr]; ok {
			continue
		}
		visited[curr] = struct{}{}

		for _, ports := range curr.Ports() {
			for _, port := range ports {
				id := port.ID
				if id == uuid.Nil {
					id = t.lookup(curr.Namespace(), port.Name)
				}

				next, ok := t.symbols[id]
				if !ok || next.Namespace() != curr.Namespace() {
					return false
				}
				stack = append(stack, next)
			}
		}
	}
	return true
}

func (t *Table) call(sb *Symbol, name string) error {
	out := port.NewOut()
	defer out.Close()

	ports := sb.Ports()
	for _, port := range ports[name] {
		id := port.ID
		if id == uuid.Nil {
			id = t.lookup(sb.Namespace(), port.Name)
		}

		if ref, ok := t.symbols[id]; ok && ref.Namespace() == sb.Namespace() {
			if in := ref.In(port.Port); in != nil {
				out.Link(in)
			}
		}
	}

	payload, err := types.Marshal(sb.Spec)
	if err != nil {
		return err
	}

	_, err = port.Send(out, payload)
	return err
}

func (t *Table) lookup(namespace, name string) uuid.UUID {
	if ns, ok := t.namespaces[namespace]; ok {
		return ns[name]
	}
	return uuid.Nil
}
