package event

import (
	"sync"

	"github.com/gofrs/uuid"
)

type Partition struct {
	queues    map[uuid.UUID]*Queue
	consumers map[uuid.UUID][]*Consumer
	mu        sync.RWMutex
}

func NewPartition() *Partition {
	return &Partition{
		queues:    make(map[uuid.UUID]*Queue),
		consumers: make(map[uuid.UUID][]*Consumer),
	}
}

func (p *Partition) Write(e *Event) {
	p.mu.RLock()
	defer p.mu.RUnlock()

	for _, v := range p.queues {
		v.Push(e)
	}
}

func (p *Partition) Consumer(id uuid.UUID) *Consumer {
	p.mu.Lock()
	defer p.mu.Unlock()

	q, ok := p.queues[id]
	if !ok {
		q = NewQueue(0)
		p.queues[id] = q
	}

	c := NewConsumer(q)
	p.consumers[id] = append(p.consumers[id], c)

	go func() {
		<-c.Done()

		p.mu.Lock()
		defer p.mu.Unlock()

		for i, v := range p.consumers[id] {
			if v == c {
				p.consumers[id] = append(p.consumers[id][:i], p.consumers[id][i+1:]...)
				if len(p.consumers[id]) == 0 {
					q := p.queues[id]
					defer q.Close()

					delete(p.consumers, id)
					delete(p.queues, id)
				}
				break
			}
		}
	}()

	return c
}

func (p *Partition) Close() {
	p.mu.Lock()
	defer p.mu.Unlock()

	for _, consumers := range p.consumers {
		for _, v := range consumers {
			v.Close()
		}
	}
	for _, v := range p.queues {
		v.Close()
	}

	p.consumers = make(map[uuid.UUID][]*Consumer)
	p.queues = make(map[uuid.UUID]*Queue)
}
