package port

import (
	"sync"

	"github.com/siyul-park/uniflow/pkg/packet"
	"github.com/siyul-park/uniflow/pkg/process"
)

// InPort represents an input port for receiving data.
type InPort struct {
	readers  map[*process.Process]*packet.Reader
	listners []Listener
	mu       sync.RWMutex
}

// NewIn creates a new InPort instance.
func NewIn() *InPort {
	return &InPort{
		readers: make(map[*process.Process]*packet.Reader),
	}
}

// Accept registers a listener for processing incoming data.
func (p *InPort) Accept(h Listener) {
	p.mu.Lock()
	defer p.mu.Unlock()

	p.listners = append(p.listners, h)
}

// Open opens the input port for a given process and returns a reader.
func (p *InPort) Open(proc *process.Process) *packet.Reader {
	p.mu.Lock()
	defer p.mu.Unlock()

	reader, ok := p.readers[proc]
	if !ok {
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

		for _, h := range p.listners {
			h := h
			go h.Accept(proc)
		}
	}

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
