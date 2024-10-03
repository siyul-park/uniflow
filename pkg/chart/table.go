package chart

import (
	"slices"
	"sync"

	"github.com/gofrs/uuid"
)

type Table struct {
	charts     map[uuid.UUID]*Chart
	namespaces map[string]map[string]uuid.UUID
	refences   map[uuid.UUID][]uuid.UUID
	mu         sync.RWMutex
}

func NewTable() *Table {
	return &Table{
		charts:     make(map[uuid.UUID]*Chart),
		namespaces: make(map[string]map[string]uuid.UUID),
		refences:   make(map[uuid.UUID][]uuid.UUID),
	}
}

func (t *Table) Insert(chrt *Chart) error {
	t.mu.Lock()
	defer t.mu.Unlock()

	if _, err := t.free(chrt.GetID()); err != nil {
		return err
	}
	return t.insert(chrt)
}

func (t *Table) Free(id uuid.UUID) (bool, error) {
	t.mu.Lock()
	defer t.mu.Unlock()

	chrt, err := t.free(id)
	if err != nil {
		return false, err
	}
	return chrt != nil, nil
}

func (t *Table) Lookup(id uuid.UUID) *Chart {
	t.mu.RLock()
	defer t.mu.RUnlock()

	return t.charts[id]
}

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
	return nil
}

func (t *Table) free(id uuid.UUID) (*Chart, error) {
	chrt, ok := t.charts[id]
	if !ok {
		return nil, nil
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

func (t *Table) links(chrt *Chart) {
	for _, spec := range chrt.GetSpecs() {
		id := t.lookup(chrt.GetNamespace(), spec.GetKind())
		if !slices.Contains(t.refences[id], chrt.GetID()) {
			t.refences[id] = append(t.refences[id], chrt.GetID())
		}
	}

	for _, ref := range t.charts {
		for _, spec := range ref.GetSpecs() {
			id := t.lookup(ref.GetNamespace(), spec.GetKind())
			if id == chrt.GetID() {
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

func (t *Table) active(chrt *Chart) bool {
	var linked []*Chart

	nexts := []*Chart{chrt}
	for len(nexts) > 0 {
		chrt := nexts[len(nexts)-1]
		ok := true
		for _, sp := range chrt.Specs {
			id := t.lookup(chrt.GetNamespace(), sp.GetKind())
			next := t.charts[id]

			if next == nil || slices.Contains(nexts, next) {
				return false
			}

			ok = slices.Contains(linked, next)
			if !ok {
				nexts = append(nexts, next)
				break
			}
		}
		if ok {
			nexts = nexts[0 : len(nexts)-1]
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
