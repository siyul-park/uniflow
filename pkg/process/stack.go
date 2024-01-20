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

	for _, head := range s.heads(key) {
		values := s.values[head]
		if values[len(values)-1] == value {
			s.values[head] = values[:len(values)-1]
			return head, true
		}
	}

	return ulid.ULID{}, false
}

// Heads returns the unique heads (keys) with non-empty stacks reachable from the given key.
func (s *Stack) Heads(key ulid.ULID) []ulid.ULID {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return s.heads(key)
}

// Clear removes the stack associated with the given key and its branches if their stacks are empty.
func (s *Stack) Clear(key ulid.ULID) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.graph.Upwards(key, func(key ulid.ULID) bool {
		for _, leaf := range s.graph.Leaves(key) {
			if len(s.values[leaf]) > 0 {
				return false
			}
		}

		delete(s.values, key)
		return true
	})
}

// Size returns the total number of elements in the stack and its branches reachable from the given key.
func (s *Stack) Size(key ulid.ULID) int {
	s.mu.RLock()
	defer s.mu.RUnlock()

	size := 0
	s.graph.Upwards(key, func(key ulid.ULID) bool {
		size += len(s.values[key])
		return true
	})

	return size
}

func (s *Stack) heads(key ulid.ULID) []ulid.ULID {
	var heads []ulid.ULID
	s.graph.Upwards(key, func(key ulid.ULID) bool {
		if len(s.values[key]) > 0 {
			heads = append(heads, key)
			return false
		}
		return true
	})

	return heads
}
