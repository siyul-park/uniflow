package pipe

import "sync"

type WritePipe[T any] struct {
	reads []*ReadPipe[T]
	mu    sync.RWMutex
}

func newWrite[T any]() *WritePipe[T] {
	return &WritePipe[T]{}
}

func (p *WritePipe[T]) Write(data T) {
	p.mu.Lock()
	defer p.mu.Unlock()

	for _, read := range p.reads {
		read.write(data)
	}
}

func (p *WritePipe[T]) Links() int {
	p.mu.Lock()
	defer p.mu.Unlock()

	return len(p.reads)
}

func (p *WritePipe[T]) Link(pipe *ReadPipe[T]) {
	p.mu.Lock()
	defer p.mu.Unlock()

	for _, read := range p.reads {
		if read == pipe {
			return
		}
	}

	p.reads = append(p.reads, pipe)

	go func() {
		select {
		case <-pipe.Done():
			p.Unlink(pipe)
		}
	}()
}

func (p *WritePipe[T]) Unlink(pipe *ReadPipe[T]) {
	p.mu.Lock()
	defer p.mu.Unlock()

	for i, read := range p.reads {
		if read == pipe {
			p.reads = append(p.reads[:i], p.reads[i+1:]...)
			return
		}
	}
}
