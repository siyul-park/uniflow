package event

import "sync"

type Event struct {
	topic string
	data  map[string]any
	mu    sync.RWMutex
}

func New(topic string, data map[string]any) *Event {
	if data == nil {
		data = make(map[string]any)
	}

	return &Event{
		topic: topic,
		data:  data,
	}
}

func (e *Event) Topic() string {
	e.mu.RLock()
	defer e.mu.RUnlock()

	return e.topic
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
