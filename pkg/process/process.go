package process

import "sync"

type Process struct {
	data      *Data
	exitHooks []ExitHook
	mu        sync.Mutex
}

var _ ExitHook = (*Process)(nil)

func New() *Process {
	return &Process{
		data: newData(),
	}
}

func (p *Process) Data() *Data {
	return p.data
}

func (p *Process) AddExitHook(h ExitHook) {
	p.mu.Lock()
	defer p.mu.Unlock()

	p.exitHooks = append(p.exitHooks, h)
}

func (p *Process) Exit(err error) {
	p.mu.Lock()
	defer p.mu.Unlock()

	for _, h := range p.exitHooks {
		h.Exit(err)
	}
	p.exitHooks = nil
}
