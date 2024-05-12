package event

import (
	"sync"
)

type Queue struct {
	in   chan *Event
	out  chan *Event
	done chan struct{}
	mu   sync.RWMutex
}

func NewQueue(capacity int) *Queue {
	q := &Queue{
		in:   make(chan *Event, capacity),
		out:  make(chan *Event, capacity),
		done: make(chan struct{}),
	}

	go func() {
		defer close(q.out)
		buffer := make([]*Event, 0, capacity)

		for {
			data, ok := <-q.in
			if !ok {
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
				case data, ok := <-q.in:
					if !ok {
						return
					}
					buffer = append(buffer, data)
				case q.out <- buffer[0]:
					buffer = buffer[1:]
				}
			}
		}
	}()

	return q
}

func (q *Queue) Push(e *Event) {
	q.mu.Lock()
	defer q.mu.Unlock()

	select {
	case <-q.done:
	default:
		q.in <- e
	}
}

func (q *Queue) Pop() <-chan *Event {
	return q.out
}

func (q *Queue) Done() <-chan struct{} {
	return q.done
}

func (q *Queue) Close() {
	q.mu.Lock()
	defer q.mu.Unlock()

	select {
	case <-q.done:
		return
	default:
	}

	close(q.done)
	close(q.in)
}
