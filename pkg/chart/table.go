package chart

import (
	"slices"
	"sync"

	"github.com/gofrs/uuid"
)

// TableOption holds configurations for a Table instance.
type TableOption struct {
	LinkHooks   []LinkHook   // LoadHooks are functions executed when symbols are loaded.
	UnlinkHooks []UnlinkHook // UnloadHooks are functions executed when symbols are unloaded.
}

// Table manages charts and their references, allowing insertion, lookup, and removal.
type Table struct {
	charts      map[uuid.UUID]*Chart
	namespaces  map[string]map[string]uuid.UUID
	refences    map[uuid.UUID][]uuid.UUID
	linkHooks   LinkHooks
	unlinkHooks UnlinkHooks
	mu          sync.RWMutex
}

// NewTable creates and returns a new Table instance with the provided options.
func NewTable(opts ...TableOption) *Table {
	var linkHooks []LinkHook
	var unlinkHooks []UnlinkHook
	for _, opt := range opts {
		linkHooks = append(linkHooks, opt.LinkHooks...)
		unlinkHooks = append(unlinkHooks, opt.UnlinkHooks...)
	}

	return &Table{
		charts:      make(map[uuid.UUID]*Chart),
		namespaces:  make(map[string]map[string]uuid.UUID),
		refences:    make(map[uuid.UUID][]uuid.UUID),
		linkHooks:   linkHooks,
		unlinkHooks: unlinkHooks,
	}
}

// Insert adds a new chart to the table, freeing the previous chart if it exists.
func (t *Table) Insert(chrt *Chart) error {
	t.mu.Lock()
	defer t.mu.Unlock()

	if _, err := t.free(chrt.GetID()); err != nil {
		return err
	}
	return t.insert(chrt)
}

// Free removes a chart from the table based on its UUID and unloads it.
func (t *Table) Free(id uuid.UUID) (bool, error) {
	t.mu.Lock()
	defer t.mu.Unlock()

	chrt, err := t.free(id)
	if err != nil {
		return false, err
	}
	return chrt != nil, nil
}

// Lookup retrieves a chart from the table based on its UUID.
func (t *Table) Lookup(id uuid.UUID) *Chart {
	t.mu.RLock()
	defer t.mu.RUnlock()

	return t.charts[id]
}

// Links returns the charts linked to the chart specified by its UUID.
func (t *Table) Links(id uuid.UUID) []*Chart {
	t.mu.RLock()
	defer t.mu.RUnlock()

	chrt, ok := t.charts[id]
	if !ok {
		return nil
	}
	return t.linked(chrt)
}

// Keys returns all IDs of charts in the table.
func (t *Table) Keys() []uuid.UUID {
	t.mu.RLock()
	defer t.mu.RUnlock()

	var ids []uuid.UUID
	for id := range t.charts {
		ids = append(ids, id)
	}
	return ids
}

// Close removes all charts from the table and unloads them.
func (t *Table) Close() error {
	t.mu.Lock()
	defer t.mu.Unlock()

	for id := range t.charts {
		if _, err := t.free(id); err != nil {
			return err
		}
	}
	return nil
}

func (t *Table) insert(chrt *Chart) error {
	t.charts[chrt.GetID()] = chrt

	ns, ok := t.namespaces[chrt.GetNamespace()]
	if !ok {
		ns = make(map[string]uuid.UUID)
		t.namespaces[chrt.GetNamespace()] = ns
	}
	ns[chrt.GetName()] = chrt.GetID()

	t.links(chrt)
	return t.load(chrt)
}

func (t *Table) free(id uuid.UUID) (*Chart, error) {
	chrt, ok := t.charts[id]
	if !ok {
		return nil, nil
	}

	if err := t.unload(chrt); err != nil {
		return nil, err
	}
	t.unlinks(chrt)

	if ns, ok := t.namespaces[chrt.GetNamespace()]; ok {
		delete(ns, chrt.GetName())
		if len(ns) == 0 {
			delete(t.namespaces, chrt.GetNamespace())
		}
	}

	delete(t.charts, id)

	return chrt, nil
}

func (t *Table) load(chrt *Chart) error {
	linked := t.linked(chrt)
	for _, sb := range linked {
		if t.active(sb) {
			if err := t.linkHooks.Link(sb); err != nil {
				return err
			}
		}
	}
	return nil
}

func (t *Table) unload(chrt *Chart) error {
	linked := t.linked(chrt)
	for i := len(linked) - 1; i >= 0; i-- {
		sb := linked[i]
		if t.active(sb) {
			if err := t.unlinkHooks.Unlink(sb); err != nil {
				return err
			}
		}
	}
	return nil
}

func (t *Table) links(chrt *Chart) {
	for _, spec := range chrt.GetSpecs() {
		id := t.lookup(chrt.GetNamespace(), spec.GetKind())
		if id != uuid.Nil && !slices.Contains(t.refences[id], chrt.GetID()) {
			t.refences[id] = append(t.refences[id], chrt.GetID())
		}
	}

	for _, ref := range t.charts {
		for _, spec := range ref.GetSpecs() {
			id := t.lookup(ref.GetNamespace(), spec.GetKind())
			if id != uuid.Nil && id == chrt.GetID() {
				if !slices.Contains(t.refences[id], ref.GetID()) {
					t.refences[id] = append(t.refences[id], ref.GetID())
				}
			}
		}
	}
}

func (t *Table) unlinks(chrt *Chart) {
	for _, spec := range chrt.GetSpecs() {
		id := t.lookup(chrt.GetNamespace(), spec.GetKind())

		refences := t.refences[id]
		for i := 0; i < len(refences); i++ {
			if refences[i] == chrt.GetID() {
				refences = append(refences[:i], refences[i+1:]...)
				i--
			}
		}

		if len(refences) > 0 {
			t.refences[id] = refences
		} else {
			delete(t.refences, id)
		}
	}

	delete(t.refences, chrt.GetID())
}

func (t *Table) linked(chrt *Chart) []*Chart {
	var linked []*Chart
	paths := []*Chart{chrt}
	for len(paths) > 0 {
		sb := paths[len(paths)-1]
		ok := true
		for _, id := range t.refences[sb.GetID()] {
			next := t.charts[id]
			ok = slices.Contains(paths, next) || slices.Contains(linked, next)
			if !ok {
				paths = append(paths, next)
				break
			}
		}
		if ok {
			paths = paths[0 : len(paths)-1]
			linked = append(linked, sb)
		}
	}
	slices.Reverse(linked)
	return linked
}

func (t *Table) active(chrt *Chart) bool {
	var linked []*Chart
	paths := []*Chart{chrt}
	for len(paths) > 0 {
		chrt := paths[len(paths)-1]
		ok := true
		for _, sp := range chrt.Specs {
			id := t.lookup(chrt.GetNamespace(), sp.GetKind())
			next := t.charts[id]

			if next == nil || slices.Contains(paths, next) {
				return false
			}

			ok = slices.Contains(linked, next)
			if !ok {
				paths = append(paths, next)
				break
			}
		}
		if ok {
			paths = paths[0 : len(paths)-1]
			linked = append(linked, chrt)
		}
	}
	return true
}

func (t *Table) lookup(namespace, name string) uuid.UUID {
	if ns, ok := t.namespaces[namespace]; ok {
		return ns[name]
	}
	return uuid.Nil
}
