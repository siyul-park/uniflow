package process

import (
	"sync"

	"github.com/oklog/ulid/v2"
)

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

func (g *Graph) Delete(stem, leaf ulid.ULID) {
	g.mu.Lock()
	defer g.mu.Unlock()

	g.stems.delete(leaf, stem)
	g.leaves.delete(stem, leaf)
}

func (g *Graph) Has(stem, leaf ulid.ULID) bool {
	g.mu.RLock()
	defer g.mu.RUnlock()

	return g.stems.has(leaf, stem) && g.leaves.has(stem, leaf)
}

func (g *Graph) Stems(leaf ulid.ULID) []ulid.ULID {
	g.mu.RLock()
	defer g.mu.RUnlock()

	if g.stems == nil {
		return nil
	}
	return g.stems[leaf]
}

func (g *Graph) Leaves(stem ulid.ULID) []ulid.ULID {
	g.mu.RLock()
	defer g.mu.RUnlock()

	if g.leaves == nil {
		return nil
	}
	return g.leaves[stem]
}

func (l links) has(key, value ulid.ULID) bool {
	for _, cur := range l[key] {
		if cur == value {
			return true
		}
	}
	return false
}

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
