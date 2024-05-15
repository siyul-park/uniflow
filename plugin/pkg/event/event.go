package event

import "sync"

type Event struct {
	data any
	done chan struct{}
	mu   sync.Mutex
}

func New(data any) *Event {
	return &Event{
		data: data,
		done: make(chan struct{}),
	}
}

func (e *Event) Data() any {
	return e.data
}

func (e *Event) Done() <-chan struct{} {
	return e.done
}

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
