package process

import (
	"context"
	"sync"
)

// Process is a processing unit that isolates data processing from others.
type Process struct {
	stack      *Stack
	heap       *Heap
	err        error
	ctx        context.Context
	done       chan struct{}
	wait       sync.WaitGroup
	closeHooks []CloseHook
	dataMutex  sync.RWMutex
	closeMutex sync.Mutex
}

// New creates a new Process.
func New() *Process {
	p := &Process{
		stack: newStack(),
		heap:  newHeap(),
		done:  make(chan struct{}),
	}

	ctx, cancel := context.WithCancel(context.Background())
	p.ctx = ctx

	go func() {
		<-p.Done()
		cancel()
	}()

	return p
}

// Stack returns a process's stack.
func (p *Process) Stack() *Stack {
	return p.stack
}

// Heap returns a process's heap.
func (p *Process) Heap() *Heap {
	return p.heap
}

// Context returns a process's context.
func (p *Process) Context() context.Context {
	return p.ctx
}

// Err returns the last error encountered by the process.
func (p *Process) Err() error {
	p.dataMutex.RLock()
	defer p.dataMutex.RUnlock()

	return p.err
}

// SetErr set the last error encountered by the process.
func (p *Process) SetErr(err error) {
	p.dataMutex.Lock()
	defer p.dataMutex.Unlock()

	p.err = err
}

// AddCloseHook add a close hook to the process.
func (p *Process) AddCloseHook(hook CloseHook) {
	p.closeMutex.Lock()
	defer p.closeMutex.Unlock()

	p.closeHooks = append(p.closeHooks, hook)
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

// Close closes the process.
func (p *Process) Close() {
	p.closeMutex.Lock()
	defer p.closeMutex.Unlock()

	select {
	case <-p.done:
		return
	default:
	}

	p.wait.Wait()
	<-p.stack.Done(nil)

	for _, h := range p.closeHooks {
		if err := h.Close(); err != nil {
			p.SetErr(err)
		}
	}

	p.closeHooks = nil
	p.heap.Close()
	p.stack.Close()
	close(p.done)
}
