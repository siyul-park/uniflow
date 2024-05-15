package event

import "sync"

type Event struct {
	data any
	mu   sync.RWMutex
}

func New(data any) *Event {
	return &Event{
		data: data,
	}
}

func (e *Event) Data() any {
	e.mu.RLock()
	defer e.mu.RUnlock()

	return e.data
}
