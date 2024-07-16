package port

import (
	"sync"

	"github.com/siyul-park/uniflow/pkg/packet"
	"github.com/siyul-park/uniflow/pkg/process"
	"github.com/siyul-park/uniflow/pkg/types"
)

// OutPort represents an output port for sending data.
type OutPort struct {
	ins       []*InPort
	writers   map[*process.Process]*packet.Writer
	listeners []Listener
	mu        sync.RWMutex
}

// Write sends the payload through OutPort, handles errors, and returns the processed result or any encountered error.
func Write(out *OutPort, payload types.Value) (types.Value, error) {
	var err error

	proc := process.New()
	defer proc.Exit(err)

	writer := out.Open(proc)
	defer writer.Close()

	outPck := packet.New(payload)
	backPck := packet.Write(writer, outPck)

	payload = backPck.Payload()

	if v, ok := payload.(types.Error); ok {
		err = v.Unwrap()
	}

	if err != nil {
		return nil, err
	}
	return payload, nil
}

// NewOut creates a new OutPort instance.
func NewOut() *OutPort {
	return &OutPort{
		writers: make(map[*process.Process]*packet.Writer),
	}
}

// Accept registers a listener to handle incoming data.
func (p *OutPort) Accept(listener Listener) {
	p.mu.Lock()
	defer p.mu.Unlock()

	p.listeners = append(p.listeners, listener)
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

// Unlink disconnects the output port from an input port.
func (p *OutPort) Unlink(in *InPort) {
	p.mu.Lock()
	defer p.mu.Unlock()

	for i, cur := range p.ins {
		if cur == in {
			p.ins = append(p.ins[:i], p.ins[i+1:]...)
			break
		}
	}
}

// Open opens the output port for a given process and returns a writer.
// If the process already has an associated writer, it returns the existing one.
// Otherwise, it creates a new writer and associates it with the process.
// It also connects the writer to all linked input ports and starts data listeners.
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

		for _, h := range p.listeners {
			h := h
			go h.Accept(proc)
		}
	}

	return writer
}

// Close closes all writers associated with the output port and clears linked input ports.
func (p *OutPort) Close() {
	p.mu.Lock()
	defer p.mu.Unlock()

	for _, writer := range p.writers {
		writer.Close()
	}
	p.writers = make(map[*process.Process]*packet.Writer)
	p.ins = nil
}
