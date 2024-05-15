package event

// Consumer represents a consumer that consumes events from a queue.
type Consumer struct {
	queue *Queue
}

// NewConsumer creates a new instance of Consumer with the given queue.
func NewConsumer(queue *Queue) *Consumer {
	return &Consumer{
		queue: queue,
	}
}

// Consume returns a channel to consume events from the consumer's queue.
func (c *Consumer) Consume() <-chan *Event {
	return c.queue.Pop()
}

// Close closes the consumer's queue.
func (c *Consumer) Close() {
	c.queue.Close()
}
