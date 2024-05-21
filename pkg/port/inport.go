package port

import (
	"sync"

	"github.com/siyul-park/uniflow/pkg/process"
)

// InPort represents an input port for receiving data.
type InPort struct {
	readers   map[*process.Process]*Reader
	initHooks []InitHook
	mu        sync.RWMutex
}

// NewIn creates a new InPort instance.
func NewIn() *InPort {
	return &InPort{
		readers: make(map[*process.Process]*Reader),
	}
}

// AddInitHook adds a handler for processing incoming data.
func (p *InPort) AddInitHook(h InitHook) {
	p.mu.Lock()
	defer p.mu.Unlock()

	p.initHooks = append(p.initHooks, h)
}

// Open opens the input port for a given process and returns a reader.
func (p *InPort) Open(proc *process.Process) *Reader {
	p.mu.Lock()
	defer p.mu.Unlock()

	reader, ok := p.readers[proc]
	if !ok {
		reader = NewReader()

		select {
		case <-proc.Done():
			reader.Close()
			return reader
		default:
		}

		p.readers[proc] = reader

		go func() {
			<-proc.Done()
			p.closeWithLock(proc)
		}()

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

	for proc := range p.readers {
		p.close(proc)
	}
}

func (p *InPort) closeWithLock(proc *process.Process) {
	p.mu.Lock()
	defer p.mu.Unlock()

	p.close(proc)
}

func (p *InPort) close(proc *process.Process) {
	if reader, ok := p.readers[proc]; ok {
		delete(p.readers, proc)
		reader.Close()
	}
}
