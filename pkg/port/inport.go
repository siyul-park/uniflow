package port

import (
	"sync"

	"github.com/siyul-park/uniflow/pkg/packet"
	"github.com/siyul-park/uniflow/pkg/process"
)

// InPort represents an input port used for receiving data.
type InPort struct {
	readers    map[*process.Process]*packet.Reader
	openHooks  OpenHooks
	closeHooks CloseHooks
	listeners  Listeners
	mu         sync.RWMutex
}

// NewIn creates and returns a new InPort instance.
func NewIn() *InPort {
	return &InPort{
		readers: make(map[*process.Process]*packet.Reader),
	}
}

// AddOpenHook adds a hook to the port if it is not already present.
func (p *InPort) AddOpenHook(hook OpenHook) bool {
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

// RemoveOpenHook removes a hook from the port if it exists.
func (p *InPort) RemoveOpenHook(hook OpenHook) bool {
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

// AddCloseHook adds a close hook to the port if it is not already present.
func (p *InPort) AddCloseHook(hook CloseHook) bool {
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

// RemoveCloseHook removes a close hook from the port if it exists.
func (p *InPort) RemoveCloseHook(hook CloseHook) bool {
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

// AddListener adds a listener to the port if it is not already registered.
func (p *InPort) AddListener(listener Listener) bool {
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

// Open prepares the input port for a given process and returns a reader.
func (p *InPort) Open(proc *process.Process) *packet.Reader {
	p.mu.Lock()

	reader, ok := p.readers[proc]
	if ok {
		p.mu.Unlock()
		return reader
	}

	reader = packet.NewReader()
	p.readers[proc] = reader

	openHooks := p.openHooks
	listeners := p.listeners

	p.mu.Unlock()

	proc.AddExitHook(process.ExitFunc(func(_ error) {
		p.mu.Lock()
		delete(p.readers, proc)
		p.mu.Unlock()

		reader.Close()
	}))

	openHooks.Open(proc)
	listeners.Accept(proc)

	return reader
}

// Close shuts down all readers associated with the input port and clears hooks, listeners, and processes close hooks.
func (p *InPort) Close() {
	p.mu.Lock()

	closeHooks := p.closeHooks
	readers := p.readers

	p.readers = make(map[*process.Process]*packet.Reader)
	p.openHooks = nil
	p.closeHooks = nil
	p.listeners = nil

	p.mu.Unlock()

	closeHooks.Close()
	for _, reader := range readers {
		reader.Close()
	}
}
