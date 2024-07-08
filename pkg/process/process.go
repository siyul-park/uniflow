package process

import (
	"context"
	"sync"
)

// Process represents an individual execution unit that manages its own data and handles termination.
type Process struct {
	parent    *Process
	data      *Data
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
	StatusRunning    Status = iota // StatusRunning indicates the process is running.
	StatusTerminated               // StatusTerminated indicates the process has terminated.
)

var _ ExitHook = (*Process)(nil)

// New creates a new Process instance with a background context and initializes an exit hook for cancellation.
func New() *Process {
	ctx, cancel := context.WithCancelCause(context.Background())
	return &Process{
		data:      newData(),
		ctx:       ctx,
		exitHooks: []ExitHook{ExitFunc(cancel)},
	}
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
	ctx, cancel := context.WithCancelCause(context.Background())
	child := &Process{
		data:   p.data.Fork(),
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

// AddExitHook adds an exit hook to be executed upon process termination.
func (p *Process) AddExitHook(h ExitHook) {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.status != StatusTerminated {
		p.exitHooks = append(p.exitHooks, h)
	}
}

// Exit terminates the process with the provided error, closes data resources,
// and executes all registered exit hooks.
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
