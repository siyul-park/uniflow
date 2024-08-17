package port

import (
	"sync"

	"github.com/siyul-park/uniflow/pkg/packet"
	"github.com/siyul-park/uniflow/pkg/process"
)

// InPort represents an input port used for receiving data.
type InPort struct {
	readers   map[*process.Process]*packet.Reader
	hooks     []Hook
	listeners []Listener
	mu        sync.RWMutex
}

// NewIn creates and returns a new InPort instance.
func NewIn() *InPort {
	return &InPort{
		readers: make(map[*process.Process]*packet.Reader),
	}
}

// AddHook adds a hook to the port if it is not already present.
func (p *InPort) AddHook(hook Hook) bool {
	p.mu.Lock()
	defer p.mu.Unlock()

	for _, h := range p.hooks {
		if h == hook {
			return false
		}
	}

	p.hooks = append(p.hooks, hook)
	return true
}

// RemoveHook removes a hook from the port if it exists.
func (p *InPort) RemoveHook(hook Hook) bool {
	p.mu.Lock()
	defer p.mu.Unlock()

	for i, h := range p.hooks {
		if h == hook {
			p.hooks = append(p.hooks[:i], p.hooks[i+1:]...)
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
// If a reader for the process already exists, it is returned. Otherwise, a new reader is created.
func (p *InPort) Open(proc *process.Process) *packet.Reader {
	p.mu.Lock()

	reader, exists := p.readers[proc]
	if exists {
		p.mu.Unlock()
		return reader
	}

	reader = packet.NewReader()
	p.readers[proc] = reader

	proc.AddExitHook(process.ExitFunc(func(_ error) {
		p.mu.Lock()
		defer p.mu.Unlock()

		delete(p.readers, proc)
		reader.Close()
	}))

	hooks := p.hooks[:]
	listeners := p.listeners[:]

	p.mu.Unlock()

	for i := len(hooks) - 1; i >= 0; i-- {
		hook := hooks[i]
		hook.Open(proc)
	}

	for _, listener := range listeners {
		listener := listener
		go listener.Accept(proc)
	}

	return reader
}

// Close shuts down all readers associated with the input port and clears hooks and listeners.
func (p *InPort) Close() {
	p.mu.Lock()
	defer p.mu.Unlock()

	for _, reader := range p.readers {
		reader.Close()
	}
	p.readers = make(map[*process.Process]*packet.Reader)
	p.hooks = nil
	p.listeners = nil
}
