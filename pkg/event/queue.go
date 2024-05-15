package event

import (
	"sync"
)

// Queue represents a queue for storing events.
type Queue struct {
	in   chan *Event
	out  chan *Event
	done chan struct{}
	mu   sync.RWMutex
}

// NewQueue creates a new instance of Queue with the given capacity.
func NewQueue(capacity int) *Queue {
	q := &Queue{
		in:   make(chan *Event, capacity),
		out:  make(chan *Event),
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
				case data = <-q.in:
					buffer = append(buffer, data)
				case q.out <- buffer[0]:
					buffer = buffer[1:]
				}
			}
		}
	}()

	return q
}

// Push pushes the given event into the queue.
// It returns true if the event is successfully pushed, otherwise false if the queue is closed.
func (q *Queue) Push(e *Event) bool {
	q.mu.Lock()
	defer q.mu.Unlock()

	select {
	case <-q.done:
		return false
	default:
	}

	q.in <- e
	return true
}

// Pop returns a channel to receive events from the queue.
func (q *Queue) Pop() <-chan *Event {
	return q.out
}

// Done returns a channel to receive a signal when the queue is done processing.
func (q *Queue) Done() <-chan struct{} {
	return q.done
}

// Close closes the queue and signals that it is done processing.
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
