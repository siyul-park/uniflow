package pipe

import "sync"

type ReadPipe[T any] struct {
	in   chan T
	out  chan T
	done chan struct{}
	mu   sync.RWMutex
}

func NewRead[T any](capacity int) *ReadPipe[T] {
	p := &ReadPipe[T]{
		in:   make(chan T, capacity),
		out:  make(chan T),
		done: make(chan struct{}),
	}

	go func() {
		defer close(p.out)
		buffer := make([]T, 0, capacity)

	loop:
		for {
			data, ok := <-p.in
			if !ok {
				break loop
			}
			select {
			case p.out <- data:
				continue
			default:
			}

			buffer = append(buffer, data)

			for len(buffer) > 0 {
				select {
				case packet, ok := <-p.in:
					if !ok {
						break loop
					}
					buffer = append(buffer, packet)
				case p.out <- buffer[0]:
					buffer = buffer[1:]
				}
			}
		}
	}()

	return p
}

func (p *ReadPipe[T]) Read() <-chan T {
	return p.out
}

func (p *ReadPipe[T]) Done() <-chan struct{} {
	return p.done
}

func (p *ReadPipe[T]) Close() {
	p.mu.Lock()
	defer p.mu.Unlock()

	select {
	case <-p.done:
		return
	default:
	}

	close(p.done)
	close(p.in)
}

func (p *ReadPipe[T]) write(data T) {
	p.mu.RLock()
	defer p.mu.RUnlock()

	select {
	case <-p.done:
	default:
		p.in <- data
	}
}
