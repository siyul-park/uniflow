package event

import "sync"

// Consumer represents a consumer for consuming events from a queue.
type Consumer struct {
	queue *Queue
	done  chan struct{}
	mu    sync.RWMutex
}

// NewConsumer creates a new Consumer instance with the given queue.
func NewConsumer(queue *Queue) *Consumer {
	return &Consumer{
		queue: queue,
		done:  make(chan struct{}),
	}
}

// Consume returns a channel to receive events from the consumer's queue.
func (c *Consumer) Consume() <-chan *Event {
	return c.queue.Pop()
}

// Done returns a channel indicating when the consumer is done.
func (c *Consumer) Done() <-chan struct{} {
	return c.done
}

// Close closes the consumer and releases associated resources.
func (c *Consumer) Close() {
	c.mu.Lock()
	defer c.mu.Unlock()

	select {
	case <-c.done:
		return
	default:
	}

	close(c.done)
}
