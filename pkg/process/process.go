package process

import "sync"

type Process struct {
	data      *Data
	status    Status
	err       error
	parent    *Process
	wait      sync.WaitGroup
	exitHooks []ExitHook
	mu        sync.RWMutex
}

type Status int

const (
	StatusRunning Status = iota
	StatusTerminated
)

var _ ExitHook = (*Process)(nil)

func New() *Process {
	return &Process{
		data: newData(),
	}
}

func (p *Process) Data() *Data {
	return p.data
}

func (p *Process) Status() Status {
	p.mu.RLock()
	defer p.mu.RUnlock()

	return p.status
}

func (p *Process) Err() error {
	p.mu.RLock()
	defer p.mu.RUnlock()

	return p.err
}

func (p *Process) Parent() *Process {
	return p.parent
}

func (p *Process) Wait() {
	p.wait.Wait()
}

func (p *Process) Fork() *Process {
	p.wait.Add(1)

	child := &Process{
		data: p.data.Fork(),
		exitHooks: []ExitHook{ExitHookFunc(func(err error) {
			p.wait.Done()
		})},
		parent: p,
	}
	p.AddExitHook(ExitHookFunc(func(err error) {
		child.Exit(err)
	}))

	return child
}

func (p *Process) AddExitHook(h ExitHook) {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.status == StatusTerminated {
		return
	}

	p.exitHooks = append(p.exitHooks, h)
}

func (p *Process) Exit(err error) {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.status == StatusTerminated {
		return
	}

	for i := len(p.exitHooks) - 1; i >= 0; i-- {
		h := p.exitHooks[i]
		h.Exit(err)
	}

	p.data.Close()

	p.status = StatusTerminated
	p.err = err
	p.exitHooks = nil
}
