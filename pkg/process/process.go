package process

import (
	"context"
	"sync"
)


// Process represents a process with its state, context, and synchronization mechanisms.
type Process struct {
	data      *Data
	status    Status
	err       error
	ctx       context.Context
	parent    *Process
	exitHooks []ExitHook
	wait      sync.WaitGroup
	mu        sync.RWMutex
}

type Status int

const (
	StatusRunning Status = iota 
	StatusTerminated            
)

// Ensure Process implements ExitHook interface.
var _ ExitHook = (*Process)(nil)

// New creates a new Process with a background context.
func New() *Process {
	return NewWithContext(context.Background())
}

// NewWithContext creates a new Process with a given context.
func NewWithContext(ctx context.Context) *Process {
	ctx, cancel := context.WithCancelCause(ctx)
	proc := &Process{
		data: newData(),
		ctx:  ctx,
	}
	proc.exitHooks = append(proc.exitHooks, ExitHookFunc(cancel))
	return proc
}

// Data returns the data associated with the process.
func (p *Process) Data() *Data {
	return p.data
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

// Context returns the context associated with the process.
func (p *Process) Context() context.Context {
	return p.ctx
}

// Parent returns the parent process.
func (p *Process) Parent() *Process {
	return p.parent
}

// Wait waits for the process to complete.
func (p *Process) Wait() {
	p.wait.Wait()
}

// Fork creates a new child process from the current process.
func (p *Process) Fork() *Process {
	p.wait.Add(1)

	ctx, cancel := context.WithCancelCause(p.ctx)

	child := &Process{
		data:   p.data.Fork(),
		ctx:    ctx,
		parent: p,
		exitHooks: []ExitHook{
			ExitHookFunc(cancel),
			ExitHookFunc(func(err error) {
				p.wait.Done()
			}),
		},
	}

	p.AddExitHook(ExitHookFunc(func(err error) {
		child.Exit(err)
	}))

	return child
}

// AddExitHook adds an exit hook to the process.
func (p *Process) AddExitHook(h ExitHook) {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.status == StatusTerminated {
		return
	}

	p.exitHooks = append(p.exitHooks, h)
}

// Exit terminates the process and calls all registered exit hooks.
func (p *Process) Exit(err error) {
	p.mu.Lock()

	if p.status == StatusTerminated {
		p.mu.Unlock()
		return
	}

	exitHooks := p.exitHooks

	p.data.Close()

	p.status = StatusTerminated
	p.err = err
	p.exitHooks = nil

	p.mu.Unlock()

	for i := len(exitHooks) - 1; i >= 0; i-- {
		h := exitHooks[i]
		h.Exit(err)
	}
}
