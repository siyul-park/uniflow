package port

import (
	"sync"

	"github.com/siyul-park/uniflow/pkg/packet"
	"github.com/siyul-park/uniflow/pkg/process"
)

// OutPort represents an output port for sending data.
type OutPort struct {
	ins        []*InPort
	writers    map[*process.Process]*packet.Writer
	openHooks  OpenHooks
	closeHooks CloseHooks
	listeners  Listeners
	mu         sync.RWMutex
}

// NewOut creates and returns a new OutPort instance.
func NewOut() *OutPort {
	return &OutPort{
		writers: make(map[*process.Process]*packet.Writer),
	}
}

// AddOpenHook adds a hook for packet processing if not already present.
func (p *OutPort) AddOpenHook(hook OpenHook) bool {
	p.mu.Lock()
	defer p.mu.Unlock()

	for _, h := range p.openHooks {
		if h == hook {
			return false
		}
	}
	p.openHooks = append(p.openHooks, hook)
	return true
}

// RemoveOpenHook removes a hook from the port if present.
func (p *OutPort) RemoveOpenHook(hook OpenHook) bool {
	p.mu.Lock()
	defer p.mu.Unlock()

	for i, h := range p.openHooks {
		if h == hook {
			p.openHooks = append(p.openHooks[:i], p.openHooks[i+1:]...)
			return true
		}
	}
	return false
}

// AddCloseHook adds a close hook to the port if not already present.
func (p *OutPort) AddCloseHook(hook CloseHook) bool {
	p.mu.Lock()
	defer p.mu.Unlock()

	for _, h := range p.closeHooks {
		if h == hook {
			return false
		}
	}
	p.closeHooks = append(p.closeHooks, hook)
	return true
}

// RemoveCloseHook removes a close hook from the port if present.
func (p *OutPort) RemoveCloseHook(hook CloseHook) bool {
	p.mu.Lock()
	defer p.mu.Unlock()

	for i, h := range p.closeHooks {
		if h == hook {
			p.closeHooks = append(p.closeHooks[:i], p.closeHooks[i+1:]...)
			return true
		}
	}
	return false
}

// AddListener registers a listener for outgoing data if not already present.
func (p *OutPort) AddListener(listener Listener) bool {
	p.mu.Lock()
	defer p.mu.Unlock()

	for _, l := range p.listeners {
		if l == listener {
			return false
		}
	}
	p.listeners = append(p.listeners, listener)
	return true
}

// Links returns the number of input ports this port is connected to.
func (p *OutPort) Links() int {
	p.mu.RLock()
	defer p.mu.RUnlock()

	return len(p.ins)
}

// Link connects this output port to an input port.
func (p *OutPort) Link(in *InPort) {
	p.mu.Lock()
	defer p.mu.Unlock()

	p.ins = append(p.ins, in)

	in.AddCloseHook(CloseHookFunc(func() {
		p.Unlink(in)
	}))
}

// Unlink disconnects this output port from an input port.
func (p *OutPort) Unlink(in *InPort) {
	p.mu.Lock()
	defer p.mu.Unlock()

	for i, cur := range p.ins {
		if cur == in {
			p.ins = append(p.ins[:i], p.ins[i+1:]...)
			break
		}
	}
}

// Open opens the output port for the given process and returns a writer.
func (p *OutPort) Open(proc *process.Process) *packet.Writer {
	p.mu.Lock()

	writer, ok := p.writers[proc]
	if ok {
		p.mu.Unlock()
		return writer
	}

	writer = packet.NewWriter()
	p.writers[proc] = writer

	ins := p.ins
	openHooks := p.openHooks
	listeners := p.listeners

	p.mu.Unlock()

	openHooks.Open(proc)
	go listeners.Accept(proc)

	proc.AddExitHook(process.ExitFunc(func(_ error) {
		p.mu.Lock()
		delete(p.writers, proc)
		p.mu.Unlock()

		writer.Close()
	}))

	for _, in := range ins {
		reader := in.Open(proc)
		writer.Link(reader)
	}

	return writer
}

// Close closes all writers and clears linked input ports, hooks, and listeners.
func (p *OutPort) Close() {
	p.mu.Lock()

	closeHooks := p.closeHooks
	writers := p.writers

	p.writers = make(map[*process.Process]*packet.Writer)
	p.ins = nil
	p.openHooks = nil
	p.closeHooks = nil
	p.listeners = nil

	p.mu.Unlock()

	closeHooks.Close()
	for _, writer := range writers {
		writer.Close()
	}
}
