package port

import (
	"sync"

	"github.com/siyul-park/uniflow/pkg/process"
)

// InPort represents an input port for receiving data.
type InPort struct {
	readers  map[*process.Process]*Reader
	handlers []Handler
	mu       sync.RWMutex
}

// OutPort represents an output port for sending data.
type OutPort struct {
	ins      []*InPort
	writers  map[*process.Process]*Writer
	handlers []Handler
	mu       sync.RWMutex
}

// NewIn creates a new InPort instance.
func NewIn() *InPort {
	return &InPort{
		readers: make(map[*process.Process]*Reader),
	}
}

// AddHandler adds a handler for processing incoming data.
func (p *InPort) AddHandler(h Handler) {
	p.mu.Lock()
	defer p.mu.Unlock()

	p.handlers = append(p.handlers, h)
}

// Open opens the input port for a given process and returns a reader.
func (p *InPort) Open(proc *process.Process) *Reader {
	p.mu.Lock()
	defer p.mu.Unlock()

	reader, ok := p.readers[proc]
	if !ok {
		reader = newReader(proc.Stack(), 2)
		p.readers[proc] = reader

		go func() {
			select {
			case <-proc.Done():
				p.closeWithLock(proc)
			case <-reader.Done():
				p.closeWithLock(proc)
			}
		}()

		for _, h := range p.handlers {
			h := h
			go h.Serve(proc)
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
		reader.Close()
		delete(p.readers, proc)
	}
}

// NewOut creates a new OutPort instance.
func NewOut() *OutPort {
	return &OutPort{
		writers: make(map[*process.Process]*Writer),
	}
}

// AddHandler adds a handler for processing outgoing data.
func (p *OutPort) AddHandler(h Handler) {
	p.mu.Lock()
	defer p.mu.Unlock()

	p.handlers = append(p.handlers, h)
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

// Unlink disconnects two pipthe output port to an input portelines.
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
			writer = newWriter(proc.Stack(), 2)
			p.writers[proc] = writer

			go func() {
				select {
				case <-proc.Done():
					p.closeWithLock(proc)
				case <-writer.Done():
					p.closeWithLock(proc)
				}
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

		for _, h := range p.handlers {
			h := h
			go h.Serve(proc)
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
		writer.Close()
		delete(p.writers, proc)
	}
}
