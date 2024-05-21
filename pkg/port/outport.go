package port

import (
	"sync"

	"github.com/siyul-park/uniflow/pkg/process"
)

// OutPort represents an output port for sending data.
type OutPort struct {
	ins       []*InPort
	writers   map[*process.Process]*Writer
	initHooks []InitHook
	mu        sync.RWMutex
}

// NewOut creates a new OutPort instance.
func NewOut() *OutPort {
	return &OutPort{
		writers: make(map[*process.Process]*Writer),
	}
}

// AddInitHook adds a handler for processing outgoing data.
func (p *OutPort) AddInitHook(h InitHook) {
	p.mu.Lock()
	defer p.mu.Unlock()

	p.initHooks = append(p.initHooks, h)
}

// Links returns the number of input ports this port is connected to.
func (p *OutPort) Links() int {
	p.mu.RLock()
	defer p.mu.RUnlock()

	return len(p.ins)
}

// Link connects the output port to an input port.
func (p *OutPort) Link(in *InPort) {
	p.mu.Lock()
	defer p.mu.Unlock()

	p.ins = append(p.ins, in)
}

// Unlink disconnects two pip the output port to an input port.
func (p *OutPort) Unlink(in *InPort) {
	p.mu.Lock()
	defer p.mu.Unlock()

	for i, cur := range p.ins {
		if cur == in {
			p.ins = append(p.ins[:i], p.ins[i+1:]...)
		}
	}
}

// Open opens the output port for a given process and returns a writer.
func (p *OutPort) Open(proc *process.Process) *Writer {
	writer, ok := func() (*Writer, bool) {
		p.mu.Lock()
		defer p.mu.Unlock()

		writer, ok := p.writers[proc]
		if !ok {
			writer = NewWriter()

			select {
			case <-proc.Done():
				writer.Close()
				return writer, true
			default:
			}

			p.writers[proc] = writer

			go func() {
				<-proc.Done()
				p.closeWithLock(proc)
			}()
		}
		return writer, ok
	}()

	if !ok {
		p.mu.RLock()
		defer p.mu.RUnlock()

		for _, in := range p.ins {
			reader := in.Open(proc)
			writer.Link(reader)
		}

		for _, h := range p.initHooks {
			h := h
			go h.Init(proc)
		}
	}

	return writer
}

// Close closes all writers associated with the output port.
func (p *OutPort) Close() {
	p.mu.Lock()
	defer p.mu.Unlock()

	for proc := range p.writers {
		p.close(proc)
	}
	p.ins = nil
}

func (p *OutPort) closeWithLock(proc *process.Process) {
	p.mu.Lock()
	defer p.mu.Unlock()

	p.close(proc)
}

func (p *OutPort) close(proc *process.Process) {
	if writer, ok := p.writers[proc]; ok {
		delete(p.writers, proc)
		writer.Close()
	}
}