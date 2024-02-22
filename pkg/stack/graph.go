package stack

import (
	"sync"
)

type Graph[T comparable] struct {
	stems  nodes[T]
	leaves nodes[T]
	heads  nodes[T]
	dones  map[T]chan struct{}
	zero   T
	mu     sync.RWMutex
}

type nodes[T comparable] map[T]edges[T]
type edges[T comparable] []T

func NewGraph[T comparable]() *Graph[T] {
	return &Graph[T]{
		stems:  make(nodes[T]),
		leaves: make(nodes[T]),
		heads:  make(nodes[T]),
		dones:  make(map[T]chan struct{}),
	}
}

func (g *Graph[T]) Has(stem, leaf T) bool {
	g.mu.RLock()
	defer g.mu.RUnlock()

	var ok bool
	g.downwards(stem, func(node T) bool {
		if ok {
			return false
		}
		if node == leaf {
			ok = true
			return false
		}
		return true
	})
	return ok
}

func (g *Graph[T]) Push(stem, leaf T) {
	g.mu.Lock()
	defer g.mu.Unlock()

	g.touch(stem)
	g.touch(leaf)

	if leaf != g.zero && stem != g.zero {
		g.stems[leaf] = g.stems[leaf].append(stem)
		g.leaves[stem] = g.leaves[stem].append(leaf)
	}
}

func (g *Graph[T]) Pop(stem, leaf T) bool {
	g.mu.Lock()
	defer g.mu.Unlock()

	for _, head := range g.heads[stem] {
		if head == leaf {
			g.remove(leaf)
			return true
		}
	}
	return false
}

func (g *Graph[T]) Clear(leaf T) {
	g.mu.Lock()
	defer g.mu.Unlock()

	g.upwards(leaf, func(node T) bool {
		if len(g.leaves[node]) > 0 {
			return false
		}
		g.remove(node)
		return true
	})
}

func (g *Graph[T]) Done(stem T) <-chan struct{} {
	g.mu.Lock()
	defer g.mu.Unlock()

	done, ok := g.dones[stem]
	if !ok {
		done = make(chan struct{})
		g.dones[stem] = done
	}

	g.done(stem)

	return done
}

func (g *Graph[T]) Close() {
	g.mu.Lock()
	defer g.mu.Unlock()

	for _, done := range g.dones {
		close(done)
	}

	g.stems = make(nodes[T])
	g.leaves = make(nodes[T])
	g.heads = make(nodes[T])
	g.dones = make(map[T]chan struct{})
}

func (g *Graph[T]) roots() []T {
	var roots []T
	for key := range g.leaves {
		if len(g.stems[key]) == 0 {
			roots = append(roots, key)
		}
	}
	return roots
}

func (g *Graph[T]) touch(node T) {
	if node == g.zero {
		return
	}
	if _, ok := g.stems[node]; !ok {
		g.stems[node] = nil
	}
	if _, ok := g.leaves[node]; !ok {
		g.leaves[node] = nil
	}
	if _, ok := g.heads[node]; !ok {
		g.heads[node] = edges[T]{node}
	}
}

func (g *Graph[T]) has(node T) bool {
	if node == g.zero {
		return len(g.stems[node])+len(g.leaves[node]) > 0
	}

	if _, ok := g.stems[node]; ok {
		return true
	}
	if _, ok := g.leaves[node]; ok {
		return true
	}
	return false
}

func (g *Graph[T]) remove(node T) {
	if node == g.zero {
		return
	}

	for cur, heads := range g.heads {
		for _, head := range heads {
			if head == node {
				heads = heads.delete(node)
				heads = heads.append(g.stems[node]...)
			}
		}
		g.heads.set(cur, heads)
	}

	for _, stem := range g.stems[node] {
		g.leaves.set(stem, g.leaves[stem].delete(node))
	}
	for _, leaf := range g.leaves[node] {
		g.stems.set(leaf, g.stems[leaf].delete(node))
	}

	delete(g.stems, node)
	delete(g.leaves, node)

	g.done(node)
	g.done(g.zero)
}

func (g *Graph[T]) done(node T) {
	if !g.has(node) {
		if done, ok := g.dones[node]; ok {
			close(done)
			delete(g.dones, node)
		}
	}
}

func (g *Graph[T]) upwards(leaf T, loop func(T) bool) {
	heads := []T{leaf}
	visits := make(map[T]struct{})

	for len(heads) > 0 {
		head := heads[0]
		heads = heads[1:]

		if _, ok := visits[head]; ok {
			continue
		}
		visits[head] = struct{}{}

		steams := g.stems[head]

		if !loop(head) {
			continue
		}

		heads = append(heads, steams...)
	}
}

func (g *Graph[T]) downwards(stem T, loop func(T) bool) {
	var heads []T
	if stem == g.zero {
		heads = append(heads, g.roots()...)
	} else {
		heads = append(heads, stem)
	}
	visits := make(map[T]struct{})

	for len(heads) > 0 {
		head := heads[0]
		heads = heads[1:]

		if _, ok := visits[head]; ok {
			continue
		}
		visits[head] = struct{}{}

		leaves := g.leaves[head]

		if !loop(head) {
			continue
		}

		heads = append(heads, leaves...)
	}
}

func (n nodes[T]) set(key T, value edges[T]) {
	if len(value) > 0 {
		n[key] = value
	} else {
		delete(n, key)
	}
}

func (e edges[T]) has(element T) bool {
	for _, v := range e {
		if v == element {
			return true
		}
	}
	return false
}

func (e edges[T]) append(elements ...T) edges[T] {
	for _, v := range elements {
		if !e.has(v) {
			e = append(e, v)
		}
	}
	return e
}

func (e edges[T]) delete(elements ...T) edges[T] {
	for _, v := range elements {
		for i, cur := range e {
			if cur == v {
				e = append(e[:i], e[i+1:]...)
				break
			}
		}
	}
	return e
}
