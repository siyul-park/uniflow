package event

import "sync"

// Broker manages producers and consumers for different topics.
type Broker struct {
	producers map[string][]*Producer
	consumers map[string][]*Consumer
	mu        sync.RWMutex
}

// NewBroker creates a new instance of Broker.
func NewBroker() *Broker {
	return &Broker{
		producers: make(map[string][]*Producer),
		consumers: make(map[string][]*Consumer),
	}
}

// Producer creates a new producer for the given topic and returns it.
func (b *Broker) Producer(topic string) *Producer {
	b.mu.Lock()
	defer b.mu.Unlock()

	queue := NewQueue(0)
	producer := NewProducer(queue)

	b.producers[topic] = append(b.producers[topic], producer)

	go func() {
		<-queue.Done()

		b.mu.Lock()
		defer b.mu.Unlock()

		for i, p := range b.producers[topic] {
			if p == producer {
				b.producers[topic] = append(b.producers[topic][:i], b.producers[topic][i+1:]...)
				return
			}
		}
	}()

	go func() {
		for event := range queue.Pop() {
			func() {
				b.mu.RLock()
				defer b.mu.RUnlock()

				var wg sync.WaitGroup
				for _, consumer := range b.consumers[topic] {
					child := New(event.Data())
					wg.Add(1)
					go func() {
						<-child.Done()
						wg.Done()
					}()

					if !consumer.queue.Push(child) {
						child.Close()
					}
				}

				go func() {
					wg.Wait()
					event.Close()
				}()
			}()
		}
	}()

	return producer
}

// Consumer creates a new consumer for the given topic and returns it.
func (b *Broker) Consumer(topic string) *Consumer {
	b.mu.Lock()
	defer b.mu.Unlock()

	queue := NewQueue(0)
	consumer := NewConsumer(queue)

	b.consumers[topic] = append(b.consumers[topic], consumer)

	go func() {
		<-queue.Done()

		b.mu.Lock()
		defer b.mu.Unlock()

		for i, c := range b.consumers[topic] {
			if c == consumer {
				b.consumers[topic] = append(b.consumers[topic][:i], b.consumers[topic][i+1:]...)
				return
			}
		}
	}()

	return consumer
}

// Close closes all producers and consumers managed by the broker.
func (b *Broker) Close() {
	b.mu.Lock()
	defer b.mu.Unlock()

	for _, producers := range b.producers {
		for _, producer := range producers {
			producer.Close()
		}
	}
	b.producers = make(map[string][]*Producer)

	for _, consumers := range b.consumers {
		for _, consumer := range consumers {
			consumer.Close()
		}
	}
	b.consumers = make(map[string][]*Consumer)
}
