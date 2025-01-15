package process

import (
	"context"
	"sync"
	"time"

	"github.com/gofrs/uuid"
)

// Process represents a unit of execution with eager, status, and lifecycle management.
type Process struct {
	parent    *Process
	id        uuid.UUID
	data      map[any]any
	status    Status
	err       error
	exitHooks ExitHooks
	done      chan struct{}
	wait      sync.WaitGroup
	mu        sync.RWMutex
}

// Status represents the process state.
type Status int

const (
	StatusRunning Status = iota
	StatusTerminated
)

var _ context.Context = (*Process)(nil)

// New creates a new process with a background context and an exit hook.
func New() *Process {
	return &Process{
		id:   uuid.Must(uuid.NewV7()),
		data: make(map[any]any),
		done: make(chan struct{}),
	}
}

// ID returns the process ID.
func (p *Process) ID() uuid.UUID {
	return p.id
}

// Keys returns all eager keys in the process.
func (p *Process) Keys() []any {
	p.mu.RLock()
	defer p.mu.RUnlock()

	keys := make([]any, 0, len(p.data))
	for key := range p.data {
		keys = append(keys, key)
	}
	if p.parent != nil {
		keys = append(keys, p.parent.Keys()...)
	}
	return keys
}

// Value gets the value associated with the key.
func (p *Process) Value(key any) any {
	p.mu.RLock()
	defer p.mu.RUnlock()

	if val, ok := p.data[key]; ok {
		return val
	}
	if p.parent != nil {
		return p.parent.Value(key)
	}
	return nil
}

// SetValue saves a value with the given key.
func (p *Process) SetValue(key, val any) {
	p.mu.Lock()
	defer p.mu.Unlock()

	p.data[key] = val
}

// RemoveValue retrieves and removes the value by key, checking the parent if not found.
func (p *Process) RemoveValue(key any) any {
	p.mu.Lock()
	defer p.mu.Unlock()

	if val, ok := p.data[key]; ok {
		delete(p.data, key)
		return val
	}
	if p.parent != nil {
		return p.parent.RemoveValue(key)
	}
	return nil
}

// Status returns the process's status.
func (p *Process) Status() Status {
	p.mu.RLock()
	defer p.mu.RUnlock()

	return p.status
}

// Err returns any error from the process.
func (p *Process) Err() error {
	p.mu.RLock()
	defer p.mu.RUnlock()

	if p.err != nil {
		return p.err
	}

	select {
	case <-p.done:
		return context.Canceled
	default:
		return nil
	}
}

// Parent returns the parent process, if any.
func (p *Process) Parent() *Process {
	return p.parent
}

// Deadline returns the time when the process will be canceled, if any.
func (p *Process) Deadline() (time.Time, bool) {
	return time.Time{}, false
}

// Done returns a channel that is closed when the process is done.
func (p *Process) Done() <-chan struct{} {
	return p.done
}

// Join waits for all child processes to complete.
func (p *Process) Join() {
	p.wait.Wait()
}

// Fork creates a child process with inherited eager and context.
func (p *Process) Fork() *Process {
	p.wait.Add(1)

	child := &Process{
		id:     uuid.Must(uuid.NewV7()),
		data:   make(map[any]any),
		parent: p,
		exitHooks: []ExitHook{
			ExitFunc(func(err error) {
				p.wait.Done()
			}),
		},
		done: make(chan struct{}),
	}
	p.AddExitHook(child)
	return child
}

// Exit terminates the process with an error, running exit hooks.
func (p *Process) Exit(err error) {
	p.mu.Lock()
	exitHooks := p.exitHooks
	if p.status != StatusTerminated {
		close(p.done)

		p.data = make(map[any]any)
		p.status = StatusTerminated
		p.err = err
		p.exitHooks = nil
	}
	p.mu.Unlock()

	exitHooks.Exit(err)
}

// AddExitHook adds an exit hook, executing immediately if terminated.
func (p *Process) AddExitHook(hook ExitHook) bool {
	p.mu.Lock()

	if p.status == StatusTerminated {
		err := p.err
		p.mu.Unlock()
		hook.Exit(err)
		return false
	}

	for _, h := range p.exitHooks {
		if h == hook {
			p.mu.Unlock()
			return false
		}
	}
	p.exitHooks = append(p.exitHooks, hook)

	p.mu.Unlock()

	return true
}
