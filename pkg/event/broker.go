package event

import (
	"sync"
)

// Broker represents a message broker responsible for managing partitions and queuing messages.
type Broker struct {
	queue      *Queue
	partitions map[string]*Partition
	mu         sync.RWMutex
}

// NewBroker creates a new Broker instance and initializes its internal queue and partitions.
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

// Producer returns a new Producer instance associated with the Broker.
func (b *Broker) Producer() *Producer {
	b.mu.RLock()
	defer b.mu.RUnlock()

	return NewProducer(b.queue)
}

// Consumer returns a new Consumer instance for the specified topic associated with the Broker.
func (b *Broker) Consumer(topic string) *Consumer {
	p := b.partition(topic)
	return p.Consumer()
}

// Close closes the Broker by closing all its partitions and the underlying message queue.
func (b *Broker) Close() {
	b.mu.Lock()
	defer b.mu.Unlock()

	for _, p := range b.partitions {
		p.Close()
	}
	b.partitions = make(map[string]*Partition)

	b.queue.Close()
}

// partition retrieves or creates a partition for the specified topic.
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
