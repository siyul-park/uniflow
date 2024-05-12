package event

type Producer struct {
	queue *Queue
}

func NewProducer(queue *Queue) *Producer {
	return &Producer{
		queue: queue,
	}
}

func (p *Producer) Send(e *Event) {
	p.queue.Push(e)
}
