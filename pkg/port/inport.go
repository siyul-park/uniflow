package port

import (
	"sync"

	"github.com/siyul-park/uniflow/pkg/packet"
	"github.com/siyul-park/uniflow/pkg/process"
)

// InPort represents an input port for receiving data.
type InPort struct {
	readers   map[*process.Process]*packet.Reader
	initHooks []InitHook
	mu        sync.RWMutex
}

// NewIn creates a new InPort instance.
func NewIn() *InPort {
	return &InPort{
		readers: make(map[*process.Process]*packet.Reader),
	}
}

// AddInitHook adds a handler for processing incoming data.
func (p *InPort) AddInitHook(h InitHook) {
	p.mu.Lock()
	defer p.mu.Unlock()

	p.initHooks = append(p.initHooks, h)
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
		proc.AtExit(process.ExitHookFunc(func(_ error) {
			p.mu.Lock()
			defer p.mu.Unlock()

			delete(p.readers, proc)
			reader.Close()
		}))

		for _, h := range p.initHooks {
			h := h
			go h.Init(proc)
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
