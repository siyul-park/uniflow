package process

import (
	"context"
	"sync"

	"github.com/gofrs/uuid"
)

// Process represents an individual execution unit with its own data and termination handling.
type Process struct {
	parent    *Process
	id        uuid.UUID
	data      map[string]any
	status    Status
	err       error
	ctx       context.Context
	exitHooks []ExitHook
	wait      sync.WaitGroup
	mu        sync.RWMutex
}

// Status represents the current state of a process.
type Status int

const (
	StatusRunning    Status = iota // Indicates the process is running.
	StatusTerminated               // Indicates the process has terminated.
)

var _ ExitHook = (*Process)(nil)

// New creates a new Process with a background context and an exit hook for cancellation.
func New() *Process {
	ctx, cancel := context.WithCancelCause(context.Background())
	return &Process{
		id:        uuid.Must(uuid.NewV7()),
		data:      make(map[string]any),
		ctx:       ctx,
		exitHooks: []ExitHook{ExitFunc(cancel)},
	}
}

// ID returns the process's id.
func (p *Process) ID() uuid.UUID {
	return p.id
}

// Load retrieves the value for the given key.
func (p *Process) Load(key string) any {
	p.mu.RLock()
	defer p.mu.RUnlock()

	return p.data[key]
}

// Store stores the value under the given key.
func (p *Process) Store(key string, val any) {
	p.mu.Lock()
	defer p.mu.Unlock()

	p.data[key] = val
}

// Delete removes the value for the given key.
func (p *Process) Delete(key string) bool {
	p.mu.Lock()
	defer p.mu.Unlock()

	if _, ok := p.data[key]; ok {
		delete(p.data, key)
		return true
	}
	return false
}

// LoadAndDelete retrieves and deletes the value for the given key.
func (p *Process) LoadAndDelete(key string) any {
	p.mu.Lock()
	defer p.mu.Unlock()

	if val, ok := p.data[key]; ok {
		delete(p.data, key)
		return val
	}

	if p.parent == nil {
		return nil
	}
	return p.parent.LoadAndDelete(key)
}

// Status returns the current status of the process.
func (p *Process) Status() Status {
	p.mu.RLock()
	defer p.mu.RUnlock()

	return p.status
}

// Err returns the error associated with the process, if any.
func (p *Process) Err() error {
	p.mu.RLock()
	defer p.mu.RUnlock()

	return p.err
}

// Context returns the process's context.
func (p *Process) Context() context.Context {
	return p.ctx
}

// Parent returns the parent process, if any.
func (p *Process) Parent() *Process {
	return p.parent
}

// Wait blocks until all child processes complete.
func (p *Process) Wait() {
	p.wait.Wait()
}

// Fork creates a new child process inheriting data and context from the current process.
func (p *Process) Fork() *Process {
	p.wait.Add(1)

	ctx, cancel := context.WithCancelCause(p.ctx)
	child := &Process{
		id:     uuid.Must(uuid.NewV7()),
		data:   make(map[string]any), // Initialize child data map
		ctx:    ctx,
		parent: p,
		exitHooks: []ExitHook{
			ExitFunc(cancel),
			ExitFunc(func(err error) { p.wait.Done() }),
		},
	}
	p.AddExitHook(child)

	return child
}

// Exit terminates the process with the provided error and runs exit hooks.
func (p *Process) Exit(err error) {
	p.mu.Lock()

	if p.status == StatusTerminated {
		p.mu.Unlock()
		return
	}

	exitHooks := p.exitHooks[:]
	p.data = make(map[string]any)
	p.status = StatusTerminated
	p.err = err
	p.exitHooks = nil

	p.mu.Unlock()

	for i := len(exitHooks) - 1; i >= 0; i-- {
		h := exitHooks[i]
		h.Exit(err)
	}
}

// AddExitHook adds an exit hook to run when the process terminates.
func (p *Process) AddExitHook(hook ExitHook) bool {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.status == StatusTerminated {
		go hook.Exit(p.err)
		return false
	}

	for _, h := range p.exitHooks {
		if h == hook {
			return false
		}
	}

	p.exitHooks = append(p.exitHooks, hook)
	return true
}
