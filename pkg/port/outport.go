package port

import (
	"sync"

	"github.com/siyul-park/uniflow/pkg/packet"
	"github.com/siyul-park/uniflow/pkg/process"
)

// OutPort represents an output port for sending data.
type OutPort struct {
	ins      []*InPort
	writers  map[*process.Process]*packet.Writer
	listners []Listener
	mu       sync.RWMutex
}

// NewOut creates a new OutPort instance.
func NewOut() *OutPort {
	return &OutPort{
		writers: make(map[*process.Process]*packet.Writer),
	}
}

// Accept registers a listener for processing incoming data.
func (p *OutPort) Accept(h Listener) {
	p.mu.Lock()
	defer p.mu.Unlock()

	p.listners = append(p.listners, h)
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
func (p *OutPort) Open(proc *process.Process) *packet.Writer {
	writer, ok := func() (*packet.Writer, bool) {
		p.mu.Lock()
		defer p.mu.Unlock()

		writer, ok := p.writers[proc]
		if !ok {
			writer = packet.NewWriter()
			if proc.Status() == process.StatusTerminated {
				writer.Close()
				return writer, true
			}

			p.writers[proc] = writer
			proc.AddExitHook(process.ExitFunc(func(_ error) {
				p.mu.Lock()
				defer p.mu.Unlock()

				delete(p.writers, proc)
				writer.Close()
			}))
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

		for _, h := range p.listners {
			h := h
			go h.Accept(proc)
		}
	}

	return writer
}

// Close closes all writers associated with the output port.
func (p *OutPort) Close() {
	p.mu.Lock()
	defer p.mu.Unlock()

	for _, writer := range p.writers {
		writer.Close()
	}
	p.writers = make(map[*process.Process]*packet.Writer)
	p.ins = nil
}
