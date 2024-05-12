package event

import (
	"sync"
)

type Broker struct {
	queue      *Queue
	partitions map[string]*Partition
	mu         sync.RWMutex
}

func NewBroker() *Broker {
	b := &Broker{
		queue:      NewQueue(0),
		partitions: make(map[string]*Partition),
	}

	go func() {
		for {
			e, ok := <-b.queue.Pop()
			if !ok {
				break
			}

			p := b.partition(e.Topic())
			p.Write(e)
		}
	}()

	return b
}

func (b *Broker) Producer() *Producer {
	b.mu.RLock()
	defer b.mu.RUnlock()

	return NewProducer(b.queue)
}

func (b *Broker) Consumer(topic string) *Consumer {
	p := b.partition(topic)
	return p.Consumer()
}

func (b *Broker) Close() {
	b.mu.Lock()
	defer b.mu.Unlock()

	for _, p := range b.partitions {
		p.Close()
	}
	b.partitions = make(map[string]*Partition)

	b.queue.Close()
}

func (b *Broker) partition(topic string) *Partition {
	b.mu.Lock()
	defer b.mu.Unlock()

	p, ok := b.partitions[topic]
	if !ok {
		p = NewPartition()
		b.partitions[topic] = p
	}

	return p
}
