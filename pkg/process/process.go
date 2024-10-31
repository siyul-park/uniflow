package process

import (
	"context"
	"sync"

	"github.com/gofrs/uuid"
)

// Process represents a unit of execution with data, status, and lifecycle management.
type Process struct {
	parent    *Process        // Parent process, if any.
	id        uuid.UUID       // Unique identifier.
	data      map[string]any  // Process data storage.
	status    Status          // Current status.
	err       error           // Execution error, if any.
	ctx       context.Context // Process context.
	exitHooks ExitHooks       // Hooks to run on exit.
	wait      sync.WaitGroup  // Manages child process completion.
	mu        sync.RWMutex    // Synchronizes access to data and status.
}

// Status represents the process state.
type Status int

const (
	StatusRunning    Status = iota // Process is active.
	StatusTerminated               // Process has ended.
)

// New creates a new process with a background context and an exit hook.
func New() *Process {
	ctx, cancel := context.WithCancelCause(context.Background())
	return &Process{
		id:        uuid.Must(uuid.NewV7()),
		data:      make(map[string]any),
		ctx:       ctx,
		exitHooks: []ExitHook{ExitFunc(cancel)},
	}
}

// ID returns the process ID.
func (p *Process) ID() uuid.UUID {
	return p.id
}

// Keys returns all data keys in the process.
func (p *Process) Keys() []string {
	p.mu.RLock()
	defer p.mu.RUnlock()

	keys := make([]string, 0, len(p.data))
	for key := range p.data {
		keys = append(keys, key)
	}
	if p.parent != nil {
		keys = append(keys, p.parent.Keys()...)
	}
	return keys
}

// Load gets the value associated with the key.
func (p *Process) Load(key string) any {
	p.mu.RLock()
	defer p.mu.RUnlock()

	if val, ok := p.data[key]; ok {
		return val
	}

	if p.parent == nil {
		return nil
	}
	return p.parent.Load(key)
}

// Store saves a value with the given key.
func (p *Process) Store(key string, val any) {
	p.mu.Lock()
	defer p.mu.Unlock()

	p.data[key] = val
}

// Delete removes the value by key.
func (p *Process) Delete(key string) bool {
	p.mu.Lock()
	defer p.mu.Unlock()

	_, exists := p.data[key]
	delete(p.data, key)
	return exists
}

// LoadAndDelete retrieves and removes the value by key, checking the parent if not found.
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

	return p.err
}

// Context returns the process context.
func (p *Process) Context() context.Context {
	return p.ctx
}

// Parent returns the parent process, if any.
func (p *Process) Parent() *Process {
	return p.parent
}

// Join waits for all child processes to complete.
func (p *Process) Join() {
	p.wait.Wait()
}

// Fork creates a child process with inherited data and context.
func (p *Process) Fork() *Process {
	p.wait.Add(1)

	ctx, cancel := context.WithCancelCause(p.ctx)
	child := &Process{
		id:     uuid.Must(uuid.NewV7()),
		data:   make(map[string]any),
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

// Exit terminates the process with an error, running exit hooks.
func (p *Process) Exit(err error) {
	p.mu.Lock()

	if p.status == StatusTerminated {
		p.mu.Unlock()
		return
	}

	exitHooks := p.exitHooks

	p.data = make(map[string]any)
	p.status = StatusTerminated
	p.err = err
	p.exitHooks = nil

	p.mu.Unlock()

	exitHooks.Exit(err)
}

// AddExitHook adds an exit hook, executing immediately if terminated.
func (p *Process) AddExitHook(hook ExitHook) bool {
	if p.Status() == StatusTerminated {
		hook.Exit(p.Err())
		return false
	}

	p.mu.Lock()
	defer p.mu.Unlock()

	for _, h := range p.exitHooks {
		if h == hook {
			return false
		}
	}
	p.exitHooks = append(p.exitHooks, hook)
	return true
}
