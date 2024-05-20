package process

import (
	"context"
	"sync"
)

// Process is a processing unit that isolates data processing from others.
type Process struct {
	heap *Heap
	ctx  context.Context
	done chan struct{}
	wait sync.WaitGroup
	mu   sync.Mutex
}

// New creates a new Process.
func New() *Process {
	h := newHeap()

	p := &Process{
		heap: h,
		done: make(chan struct{}),
	}

	ctx, cancel := context.WithCancel(context.Background())
	p.ctx = ctx

	go func() {
		<-p.Done()
		cancel()
	}()

	return p
}

// Heap returns a process's heap.
func (p *Process) Heap() *Heap {
	return p.heap
}

// Context returns a process's context.
func (p *Process) Context() context.Context {
	return p.ctx
}

// Done returns a channel that is closed when the process is closed.
func (p *Process) Done() <-chan struct{} {
	return p.done
}

// Ref acquires a lock on the process.
func (p *Process) Ref(count int) {
	p.wait.Add(count)
}

// Close closes the process.
func (p *Process) Close() {
	p.mu.Lock()
	defer p.mu.Unlock()

	select {
	case <-p.done:
		return
	default:
	}

	p.wait.Wait()
	p.heap.Close()

	close(p.done)
}
