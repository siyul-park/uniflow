package process

import (
	"sync"

	"github.com/gofrs/uuid"
)

// Process is a processing unit that isolates data processing from others.
type Process struct {
	id    uuid.UUID
	graph *Graph
	stack *Stack
	share *Share
	err   error
	done  chan struct{}
	mu    sync.RWMutex
}

// New creates a new Process.
func New() *Process {
	g := newGraph()
	s := newStack(g)

	return &Process{
		id:    uuid.Must(uuid.NewV7()),
		graph: g,
		stack: s,
		share: newShare(),
		done:  make(chan struct{}),
		mu:    sync.RWMutex{},
	}
}

// ID returns the ID of the process.
func (p *Process) ID() uuid.UUID {
	p.mu.RLock()
	defer p.mu.RUnlock()

	return p.id
}

// Graph returns a process's graph.
func (p *Process) Graph() *Graph {
	p.mu.RLock()
	defer p.mu.RUnlock()

	return p.graph
}

// Stack returns a process's stack.
func (p *Process) Stack() *Stack {
	p.mu.RLock()
	defer p.mu.RUnlock()

	return p.stack
}

// Share returns a process's share.
func (p *Process) Share() *Share {
	p.mu.RLock()
	defer p.mu.RUnlock()

	return p.share
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
	p.mu.RLock()
	defer p.mu.RUnlock()

	p.stack.Push(p.ID(), p.ID())
}

// Unlock releases the read lock on the process.
func (p *Process) Unlock() {
	p.mu.RLock()
	defer p.mu.RUnlock()

	p.stack.Pop(p.ID(), p.ID())
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

	p.stack.Close()
	p.share.Close()
	p.graph.Close()

	p.err = err
	close(p.done)
}
