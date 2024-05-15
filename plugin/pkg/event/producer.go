package event

// Producer represents a producer that produces events into a queue.
type Producer struct {
	queue *Queue
}

// NewProducer creates a new instance of Producer with the given queue.
func NewProducer(queue *Queue) *Producer {
	return &Producer{
		queue: queue,
	}
}

// Produce produces the given event into the producer's queue.
func (p *Producer) Produce(e *Event) {
	p.queue.Push(e)
}

// Close closes the producer's queue.
func (p *Producer) Close() {
	p.queue.Close()
}
