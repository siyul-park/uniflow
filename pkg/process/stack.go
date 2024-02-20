package process

import (
	"sync"

	"github.com/gofrs/uuid"
)

// Stack represents a stack data structure with associated graph-based relationships.
type Stack struct {
	graph  *Graph
	values map[uuid.UUID][]uuid.UUID
	dones  map[uuid.UUID]chan struct{}
	mu     sync.RWMutex
}

func newStack(graph *Graph) *Stack {
	return &Stack{
		graph:  graph,
		values: make(map[uuid.UUID][]uuid.UUID),
		dones:  make(map[uuid.UUID]chan struct{}),
	}
}

// Push adds a value to the stack associated with the given key.
func (s *Stack) Push(key, value uuid.UUID) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.values[key] = append(s.values[key], value)
}

// Pop removes and returns the top value from the stack associated with the given key.
func (s *Stack) Pop(key, value uuid.UUID) (uuid.UUID, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()

	for _, head := range s.heads(key) {
		values := s.values[head]
		if values[len(values)-1] == value {
			s.values[head] = values[:len(values)-1]
			s.done(head)
			return head, true
		}
	}

	return uuid.UUID{}, false
}

// Heads returns the unique heads (keys) with non-empty stacks reachable from the given key.
func (s *Stack) Heads(key uuid.UUID) []uuid.UUID {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return s.heads(key)
}

// Clear removes the stack associated with the given key and its branches if their stacks are empty.
func (s *Stack) Clear(key uuid.UUID) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.graph.Upwards(key, func(key uuid.UUID) bool {
		for _, leaf := range s.graph.Leaves(key) {
			if len(s.values[leaf]) > 0 {
				return false
			}
		}
		delete(s.values, key)
		return true
	})

	s.done(key)
}

// Size returns the total number of elements in the stack and its branches reachable from the given key.
func (s *Stack) Size(key uuid.UUID) int {
	s.mu.RLock()
	defer s.mu.RUnlock()

	size := 0
	s.graph.Upwards(key, func(key uuid.UUID) bool {
		size += len(s.values[key])
		return true
	})

	return size
}

// Close removes all values in the stack.
func (s *Stack) Close() {
	s.mu.Lock()
	defer s.mu.Unlock()

	for _, wait := range s.dones {
		close(wait)
	}

	s.values = make(map[uuid.UUID][]uuid.UUID)
	s.dones = make(map[uuid.UUID]chan struct{})
}

// Done blocks until related values in the stack are emptied.
func (s *Stack) Done(key uuid.UUID) <-chan struct{} {
	s.mu.Lock()
	defer s.mu.Unlock()

	done, ok := s.dones[key]
	if !ok {
		done = make(chan struct{})
		s.dones[key] = done
	}

	if s.leaves(key) == 0 {
		close(done)
		delete(s.dones, key)
	}

	return done
}

func (s *Stack) done(key uuid.UUID) {
	for cur, done := range s.dones {
		if s.graph.Has(cur, key) && s.leaves(cur) == 0 {
			close(done)
			delete(s.dones, cur)
		}
	}
}

func (s *Stack) heads(key uuid.UUID) []uuid.UUID {
	var heads []uuid.UUID
	s.graph.Upwards(key, func(key uuid.UUID) bool {
		if len(s.values[key]) > 0 {
			heads = append(heads, key)
			return false
		}
		return true
	})

	return heads
}

func (s *Stack) leaves(key uuid.UUID) int {
	leaves := 0

	if key == (uuid.UUID{}) {
		for _, values := range s.values {
			leaves += len(values)
		}
	} else {
		s.graph.Downwards(key, func(key uuid.UUID) bool {
			leaves += len(s.values[key])
			return true
		})
	}

	return leaves
}
