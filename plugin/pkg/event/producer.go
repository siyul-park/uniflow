package event

type Producer struct {
	queue *Queue
}

func NewProducer(queue *Queue) *Producer {
	return &Producer{
		queue: queue,
	}
}

func (p *Producer) Produce(e *Event) {
	p.queue.Push(e)
}

func (p *Producer) Close() {
	p.queue.Close()
}
