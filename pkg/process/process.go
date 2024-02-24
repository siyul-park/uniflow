package process

import (
	"sync"
)

// Process is a processing unit that isolates data processing from others.
type Process struct {
	stack *Stack
	heap  *Heap
	err   error
	done  chan struct{}
	wait  sync.WaitGroup
	mu    sync.RWMutex
}

// New creates a new Process.
func New() *Process {
	return &Process{
		stack: newStack(),
		heap:  newHeap(),
		done:  make(chan struct{}),
	}
}

// Stack returns a process's stack.
func (p *Process) Stack() *Stack {
	return p.stack
}

// Heap returns a process's heap.
func (p *Process) Heap() *Heap {
	return p.heap
}

// Err returns the last error encountered by the process.
func (p *Process) Err() error {
	p.mu.RLock()
	defer p.mu.RUnlock()

	return p.err
}

// Done returns a channel that is closed when the process is closed.
func (p *Process) Done() <-chan struct{} {
	return p.done
}

// Lock acquires a read lock on the process.
func (p *Process) Lock() {
	p.wait.Add(1)
}

// Unlock releases the read lock on the process.
func (p *Process) Unlock() {
	p.wait.Done()
}

// Exit closes the process with an optional error.
func (p *Process) Exit(err error) {
	p.mu.Lock()
	defer p.mu.Unlock()

	select {
	case <-p.done:
		return
	default:
	}

	p.wait.Wait()

	<-p.stack.Done(nil)
	p.heap.Close()

	p.err = err
	close(p.done)
}
