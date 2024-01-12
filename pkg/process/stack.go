package process

import (
	"sync"

	"github.com/oklog/ulid/v2"
)

// Stack is a data structure that manages relationships between ULIDs in a trace.
type Stack struct {
	stems  map[ulid.ULID][]ulid.ULID
	leaves map[ulid.ULID][]ulid.ULID
	stacks map[ulid.ULID][]ulid.ULID
	heads  map[ulid.ULID][]ulid.ULID
	wait   sync.RWMutex
	mu     sync.RWMutex
}

func newStack() *Stack {
	return &Stack{
		stems:  make(map[ulid.ULID][]ulid.ULID),
		leaves: make(map[ulid.ULID][]ulid.ULID),
		stacks: make(map[ulid.ULID][]ulid.ULID),
		heads:  make(map[ulid.ULID][]ulid.ULID),
	}
}

// Link establishes a relationship between two ULIDs, a stem, and a leaf.
func (s *Stack) Link(stem, leaf ulid.ULID) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if stem == leaf || s.stems == nil || s.leaves == nil {
		return
	}

	if !s.isAlreadyLinked(stem, leaf) {
		s.stems[leaf] = append(s.stems[leaf], stem)
		s.leaves[stem] = append(s.leaves[stem], leaf)
	}
}

// Unlink removes a relationship between a stem and a leaf.
func (s *Stack) Unlink(stem, leaf ulid.ULID) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if stem == leaf || s.stems == nil || s.leaves == nil {
		return
	}

	s.removeLink(s.stems, leaf, stem)
	s.removeLink(s.leaves, stem, leaf)
}

// Stems return stems of the given leaf.
func (s *Stack) Stems(leaf ulid.ULID) []ulid.ULID {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if s.stems == nil {
		return nil
	}
	return s.stems[leaf]
}

// Leaves return leaves of the given stem.
func (s *Stack) Leaves(stem ulid.ULID) []ulid.ULID {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if s.leaves == nil {
		return nil
	}
	return s.leaves[stem]
}

// Push adds a value to the stack associated with a key.
func (s *Stack) Push(key, value ulid.ULID) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.stacks == nil {
		return
	}

	s.stacks[key] = append(s.stacks[key], value)
	s.wait.RLock()
}

// Pop removes and returns the top value from the stack associated with a key.
func (s *Stack) Pop(key, value ulid.ULID) ([]ulid.ULID, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.stems == nil || s.leaves == nil || s.stacks == nil || s.heads == nil {
		return nil, false
	}

	heads := s.cleanupHeads(key)

	for i, head := range heads {
		stacks := s.stacks[head]
		if len(stacks) > 0 && stacks[len(stacks)-1] == value {
			stacks = stacks[:len(stacks)-1]

			s.stacks[head] = stacks
			if len(s.stacks[head]) == 0 {
				delete(s.stacks, head)

				heads = append(heads[:i], heads[i+1:]...)
				heads = append(heads, s.stems[head]...)

				s.heads[key] = heads
				if len(s.heads[key]) == 0 {
					delete(s.heads, key)
				}

				s.move(head)
			}

			s.wait.RUnlock()
			return heads, true
		}
	}

	return nil, false
}

// Clear removes links from the child associated with a key.
func (s *Stack) Clear(key ulid.ULID) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.stems == nil || s.leaves == nil || s.stacks == nil || s.heads == nil {
		return
	}

	heads, ok := s.heads[key]
	if !ok {
		heads = []ulid.ULID{key}
	}

	visits := map[ulid.ULID]struct{}{}
	for {
		for i, head := range heads {
			if _, ok := visits[head]; ok && len(s.leaves[head]) != 0 {
				continue
			}
			visits[head] = struct{}{}

			heads = append(heads[:i], heads[i+1:]...)
			heads = append(heads, s.stems[head]...)

			if len(s.leaves[head]) == 0 {
				for range s.stacks[head] {
					s.wait.RUnlock()
				}

				delete(s.stacks, head)
				delete(s.heads, head)

				s.move(head)
			}
		}

		next := false
		for _, head := range heads {
			if _, ok := visits[head]; !ok {
				next = true
				break
			}
		}
		if !next {
			break
		}
	}
}

// Len returns the number of values in the stack associated with a key.
func (s *Stack) Len(key ulid.ULID) int {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if s.stems == nil || s.leaves == nil || s.stacks == nil || s.heads == nil {
		return 0
	}

	heads, ok := s.heads[key]
	if !ok {
		heads = []ulid.ULID{key}
	}

	visits := map[ulid.ULID]struct{}{}
	count := 0
	for {
		for i, head := range heads {
			if _, ok := visits[head]; ok {
				continue
			}
			visits[head] = struct{}{}

			heads = append(heads[:i], heads[i+1:]...)
			heads = append(heads, s.stems[head]...)

			count += len(s.stacks[head])
		}

		next := false
		for _, head := range heads {
			if _, ok := visits[head]; !ok {
				next = true
				break
			}
		}
		if !next {
			break
		}
	}

	return count
}

// Wait blocks until the stack is empty.
func (s *Stack) Wait() {
	s.wait.Lock()
	defer s.wait.Unlock()
}

// Close releases all resources associated with the Stack.
func (s *Stack) Close() {
	s.mu.Lock()
	defer s.mu.Unlock()

	for _, stacks := range s.stacks {
		for range stacks {
			s.wait.RUnlock()
		}
	}

	s.stems = nil
	s.stacks = nil
	s.heads = nil
}

func (s *Stack) cleanupHeads(key ulid.ULID) []ulid.ULID {
	heads, ok := s.heads[key]
	if !ok {
		heads = []ulid.ULID{key}
	}

	visits := map[ulid.ULID]struct{}{}
	for {
		for i, head := range heads {
			if _, ok := visits[head]; ok && len(s.leaves[head]) != 0 {
				continue
			}
			visits[head] = struct{}{}

			if steams := s.move(head); steams != nil {
				delete(s.heads, head)

				heads = append(heads[:i], heads[i+1:]...)
				heads = append(heads, steams...)
			}
		}

		next := false
		for _, head := range heads {
			if _, ok := visits[head]; !ok {
				next = true
				break
			}
		}
		if !next {
			break
		}
	}
	if len(heads) > 0 {
		s.heads[key] = heads
	}

	return heads
}

func (s *Stack) move(head ulid.ULID) []ulid.ULID {
	if len(s.leaves[head]) > 0 || len(s.stacks[head]) > 0 {
		return nil
	}

	stems := s.stems[head]
	for _, stem := range stems {
		for j, cur := range s.leaves[stem] {
			if cur == head {
				s.leaves[stem] = append(s.leaves[stem][:j], s.leaves[stem][j+1:]...)
				if len(s.leaves[stem]) == 0 {
					delete(s.leaves, stem)
				}
			}
		}
	}
	delete(s.stems, head)

	return stems
}

func (s *Stack) removeLink(links map[ulid.ULID][]ulid.ULID, key, value ulid.ULID) {
	for i, cur := range links[key] {
		if cur == value {
			links[key] = append(links[key][:i], links[key][i+1:]...)
			if len(links[key]) == 0 {
				delete(links, key)
			}
			break
		}
	}
}

func (s *Stack) isAlreadyLinked(stem, leaf ulid.ULID) bool {
	for _, cur := range s.stems[leaf] {
		if cur == stem {
			return true
		}
	}
	return false
}
