package symbol

import (
	"slices"
	"sync"

	"github.com/gofrs/uuid"
	"github.com/siyul-park/uniflow/pkg/node"
	"github.com/siyul-park/uniflow/pkg/packet"
	"github.com/siyul-park/uniflow/pkg/port"
	"github.com/siyul-park/uniflow/pkg/process"
	"github.com/siyul-park/uniflow/pkg/scheme"
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
	scheme      *scheme.Scheme
	symbols     map[uuid.UUID]*Symbol
	namespaces  map[string]map[string]uuid.UUID
	loadHooks   []LoadHook
	unloadHooks []UnloadHook
	mu          sync.RWMutex
}

// NewTable creates a new Table instance.
func NewTable(scheme *scheme.Scheme, opts ...TableOptions) *Table {
	var loadHooks []LoadHook
	var unloadHooks []UnloadHook

	for _, opt := range opts {
		loadHooks = append(loadHooks, opt.LoadHooks...)
		unloadHooks = append(unloadHooks, opt.UnloadHooks...)
	}

	return &Table{
		scheme:      scheme,
		symbols:     make(map[uuid.UUID]*Symbol),
		namespaces:  make(map[string]map[string]uuid.UUID),
		loadHooks:   loadHooks,
		unloadHooks: unloadHooks,
	}
}

// Insert adds a new symbol to the table based on the provided spec.
func (t *Table) Insert(s spec.Spec) (*Symbol, error) {
	t.mu.Lock()
	defer t.mu.Unlock()

	sym := &Symbol{
		Spec:   s,
		linked: make(map[string][]spec.PortLocation),
	}

	if _, err := t.free(sym.ID()); err != nil {
		return nil, err
	}
	if err := t.insert(sym); err != nil {
		return nil, err
	}
	return sym, nil
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

// LookupByID retrieves a symbol from the table by its ID.
func (t *Table) LookupByID(id uuid.UUID) (*Symbol, bool) {
	t.mu.RLock()
	defer t.mu.RUnlock()

	sym, ok := t.symbols[id]
	return sym, ok
}

// LookupByName retrieves a symbol from the table by its namespace and name.
func (t *Table) LookupByName(namespace, name string) (*Symbol, bool) {
	t.mu.RLock()
	defer t.mu.RUnlock()

	sym, ok := t.symbols[t.lookup(namespace, name)]
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
		namespace, ok := t.namespaces[sym.Namespace()]
		if !ok {
			namespace = make(map[string]uuid.UUID)
			t.namespaces[sym.Namespace()] = namespace
		}
		namespace[sym.Name()] = sym.ID()
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

func (t *Table) load(sym *Symbol) error {
	linked := t.linked(sym)
	for i := 0; i < len(linked); i++ {
		sym := linked[i]
		if !t.active(sym) {
			linked = append(linked[:i], linked[i+1:]...)
			i--
		}
	}

	unlinkeds := map[*Symbol]map[string][]spec.PortLocation{}
	for _, sym := range linked {
		if sym.Node == nil {
			value, err := t.init(sym)
			if err != nil {
				return err
			}

			s, err := t.scheme.Decode(sym.Spec, value)
			if err != nil {
				return err
			}

			sym.Node, err = t.scheme.Compile(s)
			if err != nil {
				return err
			}
		}

		unlinked := map[string][]spec.PortLocation{}
		for name, locations := range sym.Links() {
			out := sym.Out(name)
			if out == nil {
				continue
			}

			for _, location := range locations {
				id := location.ID
				if id == (uuid.UUID{}) {
					id = t.lookup(sym.Namespace(), location.Name)
				}

				if ref, ok := t.symbols[id]; ok {
					if ref.Namespace() == sym.Namespace() {
						if in := ref.In(location.Port); in != nil {
							out.Link(in)
						} else {
							unlinked[name] = append(unlinked[name], location)
						}
					}
				}
			}
		}
		unlinkeds[sym] = unlinked
	}

	for sym, unlinked := range unlinkeds {
		for name, locations := range unlinked {
			out := sym.Out(name)
			if out == nil {
				continue
			}

			for _, location := range locations {
				id := location.ID
				if id == (uuid.UUID{}) {
					id = t.lookup(sym.Namespace(), location.Name)
				}

				if ref, ok := t.symbols[id]; ok {
					if ref.Namespace() == sym.Namespace() {
						if in := ref.In(location.Port); in != nil {
							out.Link(in)
						}
					}
				}
			}
		}
	}

	for _, sym := range linked {
		for _, hook := range t.loadHooks {
			if err := hook.Load(sym); err != nil {
				return err
			}
		}
	}

	return nil
}

func (t *Table) unload(sym *Symbol) error {
	linked := t.linked(sym)
	for i := 0; i < len(linked); i++ {
		sym := linked[i]
		if !t.active(sym) {
			linked = append(linked[:i], linked[i+1:]...)
			i--
		}
	}

	for i := len(linked) - 1; i >= 0; i-- {
		sym := linked[i]
		for j := len(t.unloadHooks) - 1; j >= 0; j-- {
			hook := t.unloadHooks[j]
			if err := hook.Unload(sym); err != nil {
				return err
			}
		}
	}

	for i := len(linked) - 1; i >= 0; i-- {
		sym := linked[i]
		if sym.Node != nil {
			n := sym.Node
			sym.Node = nil
			if err := n.Close(); err != nil {
				return err
			}
		}
	}

	return nil
}

func (t *Table) init(sym *Symbol) (any, error) {
	out := port.NewOut()
	defer out.Close()

	links := sym.Links()
	for _, location := range links[node.PortInit] {
		id := location.ID
		if id == (uuid.UUID{}) {
			id = t.lookup(sym.Namespace(), location.Name)
		}

		if ref, ok := t.symbols[id]; ok {
			if ref.Namespace() == sym.Namespace() {
				if in := ref.In(location.Port); in != nil {
					out.Link(in)
				}
			}
		}
	}

	proc := process.New()

	writer := out.Open(proc)
	defer writer.Close()

	outPayload, err := types.TextEncoder.Encode(sym.Spec)
	if err != nil {
		return nil, err
	}

	outPck := packet.New(outPayload)
	backPck := packet.Call(writer, outPck)

	backPayload := backPck.Payload()

	if _, ok := backPayload.(types.Error); ok {
		err = backPayload.Interface().(error)
	}

	proc.Exit(err)

	return types.InterfaceOf(backPayload), nil
}

func (t *Table) links(sym *Symbol) {
	for name, locations := range sym.Links() {
		for _, location := range locations {
			id := location.ID
			if id == (uuid.UUID{}) {
				id = t.lookup(sym.Namespace(), location.Name)
			}

			if ref, ok := t.symbols[id]; ok {
				if ref.Namespace() == sym.Namespace() {
					ref.linked[location.Port] = append(ref.linked[location.Port], spec.PortLocation{
						ID:   sym.ID(),
						Name: location.Name,
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

		for name, locations := range ref.Links() {
			for _, location := range locations {
				if (location.ID == sym.ID()) || (location.Name != "" && location.Name == sym.Name()) {
					sym.linked[location.Port] = append(sym.linked[location.Port], spec.PortLocation{
						ID:   ref.ID(),
						Name: location.Name,
						Port: name,
					})
				}
			}
		}
	}
}

func (t *Table) unlinks(sym *Symbol) {
	for name, locations := range sym.Links() {
		for _, location := range locations {
			id := location.ID
			if id == (uuid.UUID{}) {
				id = t.lookup(sym.Namespace(), location.Name)
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

		for _, locations := range sym.Links() {
			for _, location := range locations {
				id := location.ID
				if id == (uuid.UUID{}) {
					id = t.lookup(sym.Namespace(), location.Name)
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
	if namespace, ok := t.namespaces[namespace]; ok {
		return namespace[name]
	}
	return uuid.UUID{}
}
