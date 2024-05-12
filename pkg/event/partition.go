package event

import (
	"sync"
)

type Partition struct {
	queues    []*Queue
	consumers []*Consumer
	mu        sync.RWMutex
}

func NewPartition() *Partition {
	return &Partition{}
}

func (p *Partition) Write(e *Event) {
	p.mu.RLock()
	defer p.mu.RUnlock()

	for _, v := range p.queues {
		v.Push(e)
	}
}

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

func (p *Partition) Close() {
	p.mu.Lock()
	defer p.mu.Unlock()

	for _, c := range p.consumers {
		c.Close()
	}
	for _, q := range p.queues {
		q.Close()
	}

	p.consumers = nil
	p.queues = nil
}
