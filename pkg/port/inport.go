package port

import (
	"sync"

	"github.com/siyul-park/uniflow/pkg/packet"
	"github.com/siyul-park/uniflow/pkg/process"
)

// InPort represents an input port for receiving data.
type InPort struct {
	readers   map[*process.Process]*packet.Reader
	listeners []Listener
	mu        sync.RWMutex
}

// NewIn creates a new InPort instance.
func NewIn() *InPort {
	return &InPort{
		readers: make(map[*process.Process]*packet.Reader),
	}
}

// AddListener registers the listener to handle incoming data if not already registered.
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

// RemoveListener unregisters the listener so it no longer handles incoming data.
func (p *InPort) RemoveListener(listener Listener) bool {
	p.mu.Lock()
	defer p.mu.Unlock()

	for i, l := range p.listeners {
		if l == listener {
			p.listeners = append(p.listeners[:i], p.listeners[i+1:]...)
			return true
		}
	}
	return false
}

// Open opens the input port for a given process and returns a reader.
// If the process already has an associated reader, it returns the existing one.
// Otherwise, it creates a new reader and associates it with the process.
func (p *InPort) Open(proc *process.Process) *packet.Reader {
	p.mu.Lock()
	defer p.mu.Unlock()

	reader, ok := p.readers[proc]
	if ok {
		return reader
	}

	reader = packet.NewReader()

	if proc.Status() == process.StatusTerminated {
		reader.Close()
		return reader
	}

	p.readers[proc] = reader

	proc.AddExitHook(process.ExitFunc(func(_ error) {
		p.mu.Lock()
		defer p.mu.Unlock()

		delete(p.readers, proc)
		reader.Close()
	}))

	listeners := p.listeners[:]
	go func() {
		for i := len(listeners) - 1; i >= 0; i-- {
			listeners[i].Accept(proc)
		}
	}()

	return reader
}

// Close closes all readers associated with the input port.
func (p *InPort) Close() {
	p.mu.Lock()
	defer p.mu.Unlock()

	for _, reader := range p.readers {
		reader.Close()
	}
	p.readers = make(map[*process.Process]*packet.Reader)
}
