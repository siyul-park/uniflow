package event

import (
	"sync"
)

// Queue represents a buffered event queue for managing event data.
type Queue struct {
	in   chan *Event
	out  chan *Event
	done chan struct{}
	mu   sync.RWMutex
}

// NewQueue creates a new Queue instance with the specified capacity.
func NewQueue(capacity int) *Queue {
	q := &Queue{
		in:   make(chan *Event, capacity),
		out:  make(chan *Event, capacity),
		done: make(chan struct{}),
	}

	go func() {
		defer close(q.in)
		defer close(q.out)

		buffer := make([]*Event, 0, capacity)

		for {
			var data *Event
			select {
			case data = <-q.in:
			case <-q.done:
				return
			}

			select {
			case q.out <- data:
				continue
			default:
			}

			buffer = append(buffer, data)

			for len(buffer) > 0 {
				select {
				case data := <-q.in:
					buffer = append(buffer, data)
				case q.out <- buffer[0]:
					buffer = buffer[1:]
				}
			}
		}
	}()

	return q
}

// Push adds an event to the queue.
func (q *Queue) Push(e *Event) {
	q.mu.Lock()
	defer q.mu.Unlock()

	select {
	case <-q.done:
	default:
		q.in <- e
	}
}

// Pop returns a channel to receive events from the queue.
func (q *Queue) Pop() <-chan *Event {
	return q.out
}

// Done returns a channel indicating when the queue is closed.
func (q *Queue) Done() <-chan struct{} {
	return q.done
}

// Close closes the queue and releases associated resources.
func (q *Queue) Close() {
	q.mu.Lock()
	defer q.mu.Unlock()

	select {
	case <-q.done:
		return
	default:
	}

	close(q.done)
}
