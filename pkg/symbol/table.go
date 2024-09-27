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
func (t *Table) Insert(sb *Symbol) error {
	t.mu.Lock()
	defer t.mu.Unlock()

	if sb.inbounds == nil {
		sb.inbounds = make(map[string][]spec.Port)
	}

	if _, err := t.free(sb.ID()); err != nil {
		return err
	}
	if err := t.insert(sb); err != nil {
		return err
	}
	return nil
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
func (t *Table) Lookup(id uuid.UUID) (*Symbol, bool) {
	t.mu.RLock()
	defer t.mu.RUnlock()

	sb, ok := t.symbols[id]
	return sb, ok
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
			if err := t.init(sb); err != nil {
				return err
			}

			for _, hook := range t.loadHooks {
				if err := hook.Load(sb); err != nil {
					return err
				}
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
			for j := len(t.unloadHooks) - 1; j >= 0; j-- {
				hook := t.unloadHooks[j]
				if err := hook.Unload(sb); err != nil {
					return err
				}
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
					if out != nil {
						if in := ref.In(port.Port); in != nil {
							out.Link(in)
						}
					}

					ref.inbounds[port.Port] = append(ref.inbounds[port.Port], spec.Port{
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
					if out != nil {
						if in := sb.In(port.Port); in != nil {
							out.Link(in)
						}
					}

					sb.inbounds[port.Port] = append(sb.inbounds[port.Port], spec.Port{
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

			var ports []spec.Port
			for _, port := range ref.inbounds[port.Port] {
				if port.ID != sb.ID() && port.Port != name {
					ports = append(ports, port)
				}
			}

			if len(ports) > 0 {
				ref.inbounds[port.Port] = ports
			} else {
				delete(ref.inbounds, port.Port)
			}
		}
	}
}

func (t *Table) linked(sb *Symbol) []*Symbol {
	var linked []*Symbol

	nexts := []*Symbol{sb}
	for len(nexts) > 0 {
		sb := nexts[len(nexts)-1]
		ok := true
		for _, ports := range sb.inbounds {
			for _, port := range ports {
				next := t.symbols[port.ID]
				ok = slices.Contains(nexts, next) || slices.Contains(linked, next)
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
			linked = append(linked, sb)
		}
	}

	slices.Reverse(linked)
	return linked
}

func (t *Table) active(sb *Symbol) bool {
	nexts := []*Symbol{sb}
	visits := map[*Symbol]struct{}{}
	for len(nexts) > 0 {
		sb := nexts[0]
		nexts = nexts[1:]

		if _, visit := visits[sb]; visit {
			continue
		}
		visits[sb] = struct{}{}

		for _, ports := range sb.Ports() {
			for _, port := range ports {
				id := port.ID
				if id == uuid.Nil {
					id = t.lookup(sb.Namespace(), port.Name)
				}

				ref, ok := t.symbols[id]
				if !ok || ref.Namespace() != sb.Namespace() {
					return false
				}

				nexts = append(nexts, ref)
			}
		}
	}
	return true
}

func (t *Table) init(sb *Symbol) error {
	out := port.NewOut()
	defer out.Close()

	ports := sb.Ports()
	for _, port := range ports[node.PortInit] {
		id := port.ID
		if id == uuid.Nil {
			id = t.lookup(sb.Namespace(), port.Name)
		}

		if ref, ok := t.symbols[id]; ok {
			if ref.Namespace() == sb.Namespace() {
				if in := ref.In(port.Port); in != nil {
					out.Link(in)
				}
			}
		}
	}

	payload, err := types.Marshal(sb.Spec)
	if err != nil {
		return err
	}

	_, err = port.Write(out, payload)
	return err
}

func (t *Table) lookup(namespace, name string) uuid.UUID {
	if ns, ok := t.namespaces[namespace]; ok {
		return ns[name]
	}
	return uuid.Nil
}
