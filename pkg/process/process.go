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

// Exit closes the process with an optional error.
func (p *Process) Exit(err error) {
	p.mu.Lock()
	defer p.mu.Unlock()

	select {
	case <-p.done:
		return
	default:
	}

	close(p.done)

	p.err = err
}
