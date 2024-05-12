package event

import (
	"sync"
)

// Partition represents a partition responsible for distributing events to consumers.
type Partition struct {
	queue    *Queue
	producer *Producer

	queues    []*Queue
	consumers []*Consumer
	mu        sync.RWMutex
}

// NewPartition creates a new Partition instance.
func NewPartition() *Partition {
	queue := NewQueue(0)
	producer := NewProducer(queue)

	p := &Partition{
		queue:    queue,
		producer: producer,
	}

	go func() {
		for e := range p.queue.Pop() {
			func() {
				p.mu.RLock()
				defer p.mu.RUnlock()

				for _, v := range p.queues {
					v.Push(e)
				}
			}()
		}
	}()

	return p
}

// Write writes the event to all queues in the partition.
func (p *Partition) Producer() *Producer {
	p.mu.RLock()
	defer p.mu.RUnlock()

	return p.producer
}

// Consumer creates a new consumer for the partition.
func (p *Partition) Consumer() *Consumer {
	p.mu.Lock()
	defer p.mu.Unlock()

	q := NewQueue(0)
	p.queues = append(p.queues, q)

	c := NewConsumer(q)
	p.consumers = append(p.consumers, c)

	go func() {
		<-c.Done()

		p.mu.Lock()
		defer p.mu.Unlock()

		for i, v := range p.consumers {
			if v == c {
				p.consumers = append(p.consumers[:i], p.consumers[i+1:]...)
				p.queues = append(p.queues[:i], p.queues[i+1:]...)
				break
			}
		}
	}()

	return c
}

// Close closes all consumers and queues associated with the partition.
func (p *Partition) Close() {
	p.mu.Lock()
	defer p.mu.Unlock()

	p.queue.Close()

	for _, c := range p.consumers {
		c.Close()
	}
	for _, q := range p.queues {
		q.Close()
	}

	p.consumers = nil
	p.queues = nil
}
