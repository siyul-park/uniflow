package process

import (
	"sync"

	"github.com/oklog/ulid/v2"
)

// Stack represents a stack data structure with associated graph-based relationships.
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

// Push adds a value to the stack associated with the given key.
func (s *Stack) Push(key, value ulid.ULID) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.values[key] = append(s.values[key], value)
}

// Pop removes and returns the top value from the stack associated with the given key.
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

// Heads returns the unique heads (keys) with non-empty stacks reachable from the given key.
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

// Clear removes the stack associated with the given key and its branches if their stacks are empty.
func (s *Stack) Clear(key ulid.ULID) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.graph.Up(key, func(key ulid.ULID) bool {
		for _, leaf := range s.graph.Leaves(key) {
			if len(s.values[leaf]) > 0 {
				return false
			}
		}

		delete(s.values, key)
		return true
	})
}

// Has checks if the stack associated with the given key is non-empty.
func (s *Stack) Has(key ulid.ULID) bool {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return len(s.values[key]) > 0
}

// Size returns the total number of elements in the stack and its branches reachable from the given key.
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
