package event

import (
	"sync"
)

// Event represents a structured event with associated data.
type Event struct {
	data any
	wait sync.WaitGroup
	done chan struct{}
	mu   sync.Mutex
}

// New creates a new Event instance with the specified data.
func New(data any) *Event {
	return &Event{
		data: data,
	}
}

// Data returns the data associated with the event.
func (e *Event) Data() any {
	return e.data
}

// Wait increments the event's reference count by delta.
func (e *Event) Wait(delta int) {
	e.wait.Add(delta)
}

// Done returns a channel indicating when the event's reference count becomes zero.
func (e *Event) Done() <-chan struct{} {
	e.mu.Lock()
	defer e.mu.Unlock()

	if e.done == nil {
		e.done = make(chan struct{})
		go func() {
			e.wait.Wait()
			close(e.done)
		}()
	}
	return e.done
}
