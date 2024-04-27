package event

import "sync"

type Event struct {
	name string
	data map[string]any
	mu   sync.RWMutex
}

func New(name string, data map[string]any) *Event {
	if data == nil {
		data = make(map[string]any)
	}

	return &Event{
		name: name,
		data: data,
	}
}

func (e *Event) Name() string {
	e.mu.RLock()
	defer e.mu.RUnlock()

	return e.name
}

func (e *Event) Get(key string) (any, bool) {
	e.mu.RLock()
	defer e.mu.RUnlock()

	val, ok := e.data[key]
	return val, ok
}

func (e *Event) Set(key string, val any) {
	e.mu.Lock()
	defer e.mu.Unlock()

	e.data[key] = val
}

func (e *Event) Delete(key string) bool {
	e.mu.Lock()
	defer e.mu.Unlock()

	if _, ok := e.data[key]; !ok {
		return false
	}

	delete(e.data, key)
	return true
}
