package process

import (
	"sync"

	"github.com/oklog/ulid/v2"
)

type (
	// Process is a processing unit that isolates data processing with others.
	Process struct {
		id    ulid.ULID
		stack *Stack
		done  chan struct{}
		mu    sync.RWMutex
	}
)

// New creates a new Process.
func New() *Process {
	return &Process{
		id:    ulid.Make(),
		stack: NewStack(),
		done:  make(chan struct{}),
		mu:    sync.RWMutex{},
	}
}

// ID returns the ID.
func (p *Process) ID() ulid.ULID {
	p.mu.RLock()
	defer p.mu.RUnlock()

	return p.id
}

// Stack returns a Stack
func (p *Process) Stack() *Stack {
	p.mu.RLock()
	defer p.mu.RUnlock()

	return p.stack
}

// Done returns a channel that is closed when is closed.
func (p *Process) Done() <-chan struct{} {
	return p.done
}

// Close closes the Process.
func (p *Process) Close() {
	p.mu.Lock()
	defer p.mu.Unlock()

	select {
	case <-p.done:
		return
	default:
	}

	close(p.done)

	p.stack.Close()
}
