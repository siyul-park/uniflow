package event

type Consumer struct {
	queue *Queue
}

func NewConsumer(queue *Queue) *Consumer {
	return &Consumer{
		queue: queue,
	}
}

func (c *Consumer) Consume() <-chan *Event {
	return c.queue.Pop()
}

func (c *Consumer) Close() {
	c.queue.Close()
}
