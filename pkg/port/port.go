package port

import (
	"sync"

	"github.com/siyul-park/uniflow/pkg/process"
)

type InPort struct {
	readers  map[*process.Process]*Reader
	handlers []Handler
	mu       sync.RWMutex
}

func NewIn() *InPort {
	return &InPort{
		readers: make(map[*process.Process]*Reader),
	}
}

func (p *InPort) AddHandler(h Handler) {
	p.mu.Lock()
	defer p.mu.Unlock()

	p.handlers = append(p.handlers, h)
}

func (p *InPort) Open(proc *process.Process) *Reader {
	reader, ok := func() (*Reader, bool) {
		p.mu.Lock()
		defer p.mu.Unlock()

		reader, ok := p.readers[proc]
		if !ok {
			reader = newReader(proc.Stack(), 0)
			p.readers[proc] = reader

			go func() {
				select {
				case <-proc.Done():
					p.closeWithLock(proc)
				case <-reader.Done():
					p.closeWithLock(proc)
				}
			}()
		}
		return reader, ok
	}()

	if !ok {
		p.mu.RLock()
		defer p.mu.RUnlock()

		for _, h := range p.handlers {
			h := h
			go h.Serve(proc)
		}
	}

	return reader
}

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

type OutPort struct {
	ins      []*InPort
	writers  map[*process.Process]*Writer
	handlers []Handler
	mu       sync.RWMutex
}

func NewOut() *OutPort {
	return &OutPort{
		writers: make(map[*process.Process]*Writer),
	}
}

func (p *OutPort) AddHandler(h Handler) {
	p.mu.Lock()
	defer p.mu.Unlock()

	p.handlers = append(p.handlers, h)
}

func (p *OutPort) Links() int {
	p.mu.RLock()
	defer p.mu.RUnlock()

	return len(p.ins)
}

func (p *OutPort) Link(in *InPort) {
	p.mu.Lock()
	defer p.mu.Unlock()

	p.ins = append(p.ins, in)
}

func (p *OutPort) Open(proc *process.Process) *Writer {
	writer, ok := func() (*Writer, bool) {
		p.mu.Lock()
		defer p.mu.Unlock()

		writer, ok := p.writers[proc]
		if !ok {
			writer = newWriter(proc.Stack(), 0)
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
