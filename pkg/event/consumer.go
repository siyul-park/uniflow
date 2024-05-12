package event

import "sync"

type Consumer struct {
	queue *Queue
	done  chan struct{}
	mu    sync.RWMutex
}

func NewConsumer(queue *Queue) *Consumer {
	return &Consumer{
		queue: queue,
		done:  make(chan struct{}),
	}
}

func (c *Consumer) Consume() <-chan *Event {
	return c.queue.Pop()
}

func (c *Consumer) Done() <-chan struct{} {
	return c.done
}

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
