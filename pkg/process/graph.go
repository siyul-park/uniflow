package process

import (
	"sync"

	"github.com/oklog/ulid/v2"
)

// Graph represents a directed acyclic graph with stems and leaves.
type Graph struct {
	stems  links
	leaves links
	mu     sync.RWMutex
}

type links map[ulid.ULID][]ulid.ULID

func newGraph() *Graph {
	return &Graph{
		stems:  make(links),
		leaves: make(links),
	}
}

// Add creates a directed edge from stem to leaf in the graph.
func (g *Graph) Add(stem, leaf ulid.ULID) {
	g.mu.Lock()
	defer g.mu.Unlock()

	if !g.stems.has(leaf, stem) {
		g.stems[leaf] = append(g.stems[leaf], stem)
	}
	if !g.leaves.has(stem, leaf) {
		g.leaves[stem] = append(g.leaves[stem], leaf)
	}
}

// Delete removes the directed edge from stem to leaf in the graph.
func (g *Graph) Delete(stem, leaf ulid.ULID) {
	g.mu.Lock()
	defer g.mu.Unlock()

	g.stems.delete(leaf, stem)
	g.leaves.delete(stem, leaf)
}

// Has checks if there is a directed path from stem to leaf in the graph.
func (g *Graph) Has(stem, leaf ulid.ULID) bool {
	var ok bool
	g.Up(leaf, func(key ulid.ULID) bool {
		if ok {
			return false
		}
		if key == stem {
			ok = true
			return false
		}
		return true
	})
	return ok
}

// Stems returns the stems associated with the given leaf in the graph.
func (g *Graph) Stems(leaf ulid.ULID) []ulid.ULID {
	g.mu.RLock()
	defer g.mu.RUnlock()

	if g.stems == nil {
		return nil
	}
	return g.stems[leaf]
}

// Leaves returns the leaves associated with the given stem in the graph.
func (g *Graph) Leaves(stem ulid.ULID) []ulid.ULID {
	g.mu.RLock()
	defer g.mu.RUnlock()

	if g.leaves == nil {
		return nil
	}
	return g.leaves[stem]
}

// Up traverses the graph upwards from the specified leaf, invoking the provided function on each visited node.
func (g *Graph) Up(leaf ulid.ULID, f func(ulid.ULID) bool) {
	g.mu.RLock()
	defer g.mu.RUnlock()

	heads := []ulid.ULID{leaf}
	visits := make(map[ulid.ULID]struct{})

	for len(heads) > 0 {
		head := heads[0]
		heads = heads[1:]

		if _, ok := visits[head]; ok {
			continue
		}
		visits[head] = struct{}{}

		if !f(head) {
			continue
		}

		heads = append(heads, g.stems[head]...)
	}
}

// has checks if the given key is associated with the specified value in the links.
func (l links) has(key, value ulid.ULID) bool {
	for _, cur := range l[key] {
		if cur == value {
			return true
		}
	}
	return false
}

// delete removes the specified value from the links associated with the given key.
func (l links) delete(key, value ulid.ULID) {
	for i, cur := range l[key] {
		if cur == value {
			l[key] = append(l[key][:i], l[key][i+1:]...)
			if len(l[key]) == 0 {
				delete(l, key)
			}
			break
		}
	}
}
