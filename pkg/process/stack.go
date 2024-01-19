package process

import (
	"sync"

	"github.com/oklog/ulid/v2"
)

type Stack struct {
	graph  *Graph
	values map[ulid.ULID][]ulid.ULID
	mu     sync.RWMutex
}

func newStack(graph *Graph) *Stack {
	return &Stack{
		graph:  graph,
		values: make(map[ulid.ULID][]ulid.ULID),
	}
}

func (s *Stack) Push(key, value ulid.ULID) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.values[key] = append(s.values[key], value)
}

func (s *Stack) Pop(key, value ulid.ULID) (ulid.ULID, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()

	var head ulid.ULID
	var ok bool
	s.graph.Up(key, func(key ulid.ULID) bool {
		if ok {
			return false
		}

		values := s.values[key]
		if len(values) == 0 {
			return true
		}

		if values[len(values)-1] == value {
			s.values[key] = values[:len(values)-1]

			head = key
			ok = true

			return false
		}
		return true
	})

	return head, ok
}

func (s *Stack) Heads(key ulid.ULID) []ulid.ULID {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var heads []ulid.ULID
	s.graph.Up(key, func(key ulid.ULID) bool {
		if len(s.values[key]) > 0 {
			heads = append(heads, key)
			return false
		}
		return true
	})

	return heads
}

func (s *Stack) Clear(key ulid.ULID) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.graph.Up(key, func(key ulid.ULID) bool {
		for _, leave := range s.graph.Leaves(key) {
			if len(s.values[leave]) > 0 {
				return false
			}
		}

		delete(s.values, key)
		return true
	})
}

func (s *Stack) Has(key ulid.ULID) bool {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return len(s.values[key]) > 0
}

func (s *Stack) Size(key ulid.ULID) int {
	s.mu.RLock()
	defer s.mu.RUnlock()

	size := 0
	s.graph.Up(key, func(key ulid.ULID) bool {
		size += len(s.values[key])
		return true
	})

	return size
}
