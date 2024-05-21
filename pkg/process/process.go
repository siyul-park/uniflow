package process

import "sync"

type Process struct {
	status    Status
	err       error
	data      *Data
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

func (p *Process) Status() Status {
	p.mu.RLock()
	defer p.mu.RUnlock()

	return p.status
}

func (p *Process) Error() error {
	p.mu.RLock()
	defer p.mu.RUnlock()

	return p.err
}

func (p *Process) Data() *Data {
	return p.data
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
