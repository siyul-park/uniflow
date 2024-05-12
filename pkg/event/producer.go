package event

// Producer represents a producer responsible for producing events to a queue.
type Producer struct {
	queue *Queue
}

// NewProducer creates a new Producer instance with the given queue.
func NewProducer(queue *Queue) *Producer {
	return &Producer{
		queue: queue,
	}
}

// Produce produces the event to the producer's queue.
func (p *Producer) Produce(e *Event) {
	e.Wait(1)
	if !p.queue.Push(e) {
		e.Wait(-1)
	}
}
