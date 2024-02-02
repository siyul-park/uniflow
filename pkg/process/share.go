package process

import "sync"

// Share is a container of common resources that are shared during the process life cycle.
type Share struct {
	data map[string]any
	mu   sync.RWMutex
}

func newShare() *Share {
	return &Share{
		data: make(map[string]any),
	}
}

// Load returns the value stored in the map for a key, or nil if no value is present.
func (s *Share) Load(key string) any {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return s.data[key]
}

// Store sets the value for a key.
func (s *Share) Store(key string, val any) bool {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, ok := s.data[key]; ok {
		return false
	}
	s.data[key] = val
	return true
}

// Delete deletes the value for a key.
func (s *Share) Delete(key string) bool {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, ok := s.data[key]; !ok {
		return false
	}
	delete(s.data, key)
	return true
}

// Close remove all value.
func (s *Share) Close() {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.data = make(map[string]any)
}
