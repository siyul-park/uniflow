package store

import (
	"sync"
)

type Source interface {
	Open(name string) (Store, error)
	Close() error
}

type source struct {
	stores map[string]Store
	mu     sync.Mutex
}

var _ Source = (*source)(nil)

func NewSource() Source {
	return &source{stores: make(map[string]Store)}
}

func (s *source) Open(name string) (Store, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	v, ok := s.stores[name]
	if !ok {
		v = New()
		s.stores[name] = v
	}
	return v, nil
}

func (s *source) Close() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.stores = make(map[string]Store)
	return nil
}
