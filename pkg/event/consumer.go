package event

type Consumer struct {
	queue *Queue
}

func NewConsumer(queue *Queue) *Consumer {
	return &Consumer{
		queue: queue,
	}
}

func (c *Consumer) Read() <-chan *Event {
	return c.queue.Pop()
}
