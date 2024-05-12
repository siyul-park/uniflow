package event

import (
	"sync"
)

// Event represents a structured event with associated data.
type Event struct {
	data any
	mu   sync.RWMutex
}

// New creates a new Event instance with the specified data.
func New(data any) *Event {
	return &Event{
		data: data,
	}
}

// Data returns the data associated with the event.
func (e *Event) Data() any {
	e.mu.RLock()
	defer e.mu.RUnlock()

	return e.data
}
