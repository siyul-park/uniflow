package process

import (
	"math"
	"slices"
	"sync"

	"github.com/siyul-park/uniflow/pkg/packet"
)

type Stack struct {
	stems  nodes
	leaves nodes
	heads  nodes
	dones  map[*packet.Packet]chan struct{}
	mu     sync.RWMutex
}

type nodes map[*packet.Packet]edges
type edges []*packet.Packet

func newStack() *Stack {
	return &Stack{
		stems:  make(nodes),
		leaves: make(nodes),
		heads:  make(nodes),
		dones:  make(map[*packet.Packet]chan struct{}),
	}
}

func (s *Stack) Has(stem, leaf *packet.Packet) bool {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if stem == nil {
		_, ok := s.stems[leaf]
		return ok
	}

	var ok bool
	s.downwards(stem, func(node *packet.Packet, _ []*packet.Packet) bool {
		if node == leaf {
			ok = true
		}
		return !ok
	})
	return ok
}

func (s *Stack) Add(stem, leaf *packet.Packet) {
	s.mu.Lock()
	defer s.mu.Unlock()

	delete(s.leaves, nil)

	s.touch(stem)
	s.touch(leaf)

	if stem != leaf && leaf != nil && stem != nil {
		s.stems[leaf] = s.stems[leaf].append(stem)
		s.leaves[stem] = s.leaves[stem].append(leaf)
	}

	s.refreshRoot(stem)
	s.refreshRoot(leaf)
}

func (s *Stack) Unwind(leaf, stem *packet.Packet) bool {
	s.mu.Lock()
	defer s.mu.Unlock()

	path := s.path(stem, leaf)
	if len(path) == 0 {
		return false
	}

	for _, node := range path {
		s.remove(node)
	}
	return true
}

func (s *Stack) Clear(leaf *packet.Packet) {
	s.mu.Lock()
	defer s.mu.Unlock()

	for _, head := range s.heads[leaf] {
		s.upwards(head, func(node *packet.Packet, _ []*packet.Packet) bool {
			if len(s.leaves[node]) > 0 {
				return false
			}
			s.remove(node)
			return true
		})
	}
}

func (s *Stack) Cost(stem, leaf *packet.Packet) int {
	s.mu.RLock()
	defer s.mu.RUnlock()

	cost := len(s.path(stem, leaf)) - 1
	if cost < 0 {
		return math.MaxInt
	}
	return cost
}

func (s *Stack) Done(stem *packet.Packet) <-chan struct{} {
	s.mu.Lock()
	defer s.mu.Unlock()

	done, ok := s.dones[stem]
	if !ok {
		done = make(chan struct{})
		s.dones[stem] = done
	}

	s.done(stem)

	return done
}

func (s *Stack) Close() {
	s.mu.Lock()
	defer s.mu.Unlock()

	for _, done := range s.dones {
		close(done)
	}

	s.stems = make(nodes)
	s.leaves = make(nodes)
	s.heads = make(nodes)
	s.dones = make(map[*packet.Packet]chan struct{})
}

func (s *Stack) path(stem, leaf *packet.Packet) []*packet.Packet {
	var path []*packet.Packet
	for _, head := range s.heads[leaf] {
		s.upwards(head, func(node *packet.Packet, cur []*packet.Packet) bool {
			if node == stem {
				path = cur
			}
			return path == nil
		})
	}
	slices.Reverse(path)
	return path
}

func (s *Stack) touch(node *packet.Packet) {
	if node == nil {
		return
	}

	if _, ok := s.stems[node]; !ok {
		s.stems[node] = nil
	}
	if _, ok := s.leaves[node]; !ok {
		s.leaves[node] = nil
	}

	if _, ok := s.heads[node]; !ok {
		s.heads[node] = edges{node}
	} else {
		for _, head := range s.heads[node] {
			var ok bool
			s.downwards(node, func(node *packet.Packet, cur []*packet.Packet) bool {
				if node == head {
					ok = true
				}
				return !ok
			})

			if !ok {
				s.stems[node] = s.stems[node].append(head)
				s.leaves[head] = s.leaves[head].append(node)
			}
		}
	}
}

func (s *Stack) has(node *packet.Packet) bool {
	if node == nil {
		return len(s.stems[node]) > 0
	}
	_, ok := s.stems[node]
	return ok
}

func (s *Stack) remove(node *packet.Packet) {
	for cur, heads := range s.heads {
		if heads.has(node) {
			heads = heads.delete(node)
			heads = heads.append(s.stems[node]...)
		}
		if len(heads) > 0 {
			s.heads[cur] = heads
		} else {
			delete(s.heads, cur)
		}
	}

	for _, stem := range s.stems[node] {
		s.leaves[stem] = s.leaves[stem].delete(node)
	}
	for _, leaf := range s.leaves[node] {
		s.stems[leaf] = s.stems[leaf].delete(node)
	}

	if s.leaves[nil].has(node) {
		for _, leaf := range s.leaves[node] {
			s.refreshRoot(leaf)
		}
	}

	delete(s.stems, node)
	delete(s.leaves, node)

	s.done(node)
	s.done(nil)
}

func (s *Stack) refreshRoot(leaf *packet.Packet) {
	if leaf == nil {
		return
	}
	if len(s.stems[leaf]) > 0 {
		s.leaves[nil] = s.leaves[nil].delete(leaf)
	} else {
		s.leaves[nil] = s.leaves[nil].append(leaf)
	}
}

func (s *Stack) done(node *packet.Packet) {
	if !s.has(node) {
		if done, ok := s.dones[node]; ok {
			close(done)
			delete(s.dones, node)
		}
	}
}

func (s *Stack) upwards(leaf *packet.Packet, loop func(*packet.Packet, []*packet.Packet) bool) {
	heads := []*packet.Packet{leaf}
	parents := make(map[*packet.Packet]*packet.Packet)
	visits := make(map[*packet.Packet]struct{})
	for len(heads) > 0 {
		head := heads[0]
		heads = heads[1:]

		if _, ok := visits[head]; ok {
			continue
		}
		visits[head] = struct{}{}

		stems := s.stems[head]
		for _, stem := range stems {
			parents[stem] = head
		}

		path := []*packet.Packet{head}
		for {
			if parent, ok := parents[path[len(path)-1]]; ok {
				path = append(path, parent)
			} else {
				break
			}
		}

		if !loop(head, path) {
			continue
		}

		heads = append(heads, stems...)
	}
}

func (s *Stack) downwards(stem *packet.Packet, loop func(*packet.Packet, []*packet.Packet) bool) {
	var heads []*packet.Packet
	if stem == nil {
		heads = append(heads, s.leaves[nil]...)
	} else {
		heads = append(heads, stem)
	}
	parents := make(map[*packet.Packet]*packet.Packet)
	visits := make(map[*packet.Packet]struct{})
	for len(heads) > 0 {
		head := heads[0]
		heads = heads[1:]

		if _, ok := visits[head]; ok {
			continue
		}
		visits[head] = struct{}{}

		leaves := s.leaves[head]
		for _, leaf := range leaves {
			parents[leaf] = head
		}

		path := []*packet.Packet{head}
		for {
			if parent, ok := parents[path[len(path)-1]]; ok {
				path = append(path, parent)
			} else {
				break
			}
		}

		if !loop(head, path) {
			continue
		}

		heads = append(heads, leaves...)
	}
}

func (e edges) has(element *packet.Packet) bool {
	for _, v := range e {
		if v == element {
			return true
		}
	}
	return false
}

func (e edges) append(elements ...*packet.Packet) edges {
	for _, v := range elements {
		if !e.has(v) {
			e = append(e, v)
		}
	}
	return e
}

func (e edges) delete(elements ...*packet.Packet) edges {
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
