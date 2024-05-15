package event

import "sync"

// Event represents an event with associated data.
type Event struct {
	data any
	done chan struct{}
	mu   sync.Mutex
}

// New creates a new instance of Event with the given data.
func New(data any) *Event {
	return &Event{
		data: data,
		done: make(chan struct{}),
	}
}

// Data returns the data associated with the event.
func (e *Event) Data() any {
	return e.data
}

// Done returns a channel to receive a signal when the event is done processing.
func (e *Event) Done() <-chan struct{} {
	return e.done
}

// Close closes the event and signals that it is done processing.
func (e *Event) Close() {
	e.mu.Lock()
	defer e.mu.Unlock()

	select {
	case <-e.done:
		return
	default:
	}

	close(e.done)
}
