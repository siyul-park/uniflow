package event

import "sync"

type Broker struct {
	producers map[string][]*Producer
	consumers map[string][]*Consumer
	mu        sync.RWMutex
}

func NewBroker() *Broker {
	return &Broker{
		producers: make(map[string][]*Producer),
		consumers: make(map[string][]*Consumer),
	}
}

func (b *Broker) Producer(topic string) *Producer {
	b.mu.Lock()
	defer b.mu.Unlock()

	q := NewQueue(0)
	p := NewProducer(q)

	b.producers[topic] = append(b.producers[topic], p)

	go func() {
		<-q.Done()

		b.mu.Lock()
		defer b.mu.Unlock()

		for i, producer := range b.producers[topic] {
			if producer == p {
				b.producers[topic] = append(b.producers[topic][:i], b.producers[topic][i+1:]...)
				return
			}
		}
	}()

	go func() {
		for e := range q.Pop() {
			func() {
				b.mu.RLock()
				defer b.mu.RUnlock()

				wg := sync.WaitGroup{}
				for _, consumer := range b.consumers[topic] {
					child := New(e.Data())

					wg.Add(1)
					go func() {
						<-child.Done()
						wg.Add(-1)
					}()

					if !consumer.queue.Push(child) {
						child.Close()
					}
				}

				go func() {
					wg.Wait()
					e.Close()
				}()
			}()
		}
	}()

	return p
}

func (b *Broker) Consumer(topic string) *Consumer {
	b.mu.Lock()
	defer b.mu.Unlock()

	q := NewQueue(0)
	c := NewConsumer(q)

	b.consumers[topic] = append(b.consumers[topic], c)

	go func() {
		<-q.Done()

		b.mu.Lock()
		defer b.mu.Unlock()

		for i, consumer := range b.consumers[topic] {
			if consumer == c {
				b.consumers[topic] = append(b.consumers[topic][:i], b.consumers[topic][i+1:]...)
				return
			}
		}
	}()

	return c
}

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
