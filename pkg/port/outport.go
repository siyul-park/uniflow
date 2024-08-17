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
	hooks     []Hook
	listeners []Listener
	mu        sync.RWMutex
}

// Write sends the payload through the OutPort and returns the result or an error.
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

// NewOut creates and returns a new OutPort instance.
func NewOut() *OutPort {
	return &OutPort{
		writers: make(map[*process.Process]*packet.Writer),
	}
}

// AddHook adds a hook for packet processing if not already present.
func (p *OutPort) AddHook(hook Hook) bool {
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

// RemoveHook removes a hook from the port if present.
func (p *OutPort) RemoveHook(hook Hook) bool {
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

// AddListener registers a listener for outgoing data if not already present.
func (p *OutPort) AddListener(listener Listener) bool {
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

// Links returns the number of input ports this port is connected to.
func (p *OutPort) Links() int {
	p.mu.RLock()
	defer p.mu.RUnlock()

	return len(p.ins)
}

// Link connects this output port to an input port.
func (p *OutPort) Link(in *InPort) {
	p.mu.Lock()
	defer p.mu.Unlock()

	p.ins = append(p.ins, in)
}

// Unlink disconnects this output port from an input port.
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

// Open opens the output port for the given process and returns a writer.
// It connects the writer to all linked input ports and starts data listeners.
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

		for _, in := range p.ins {
			reader := in.Open(proc)
			writer.Link(reader)
		}

		hooks := p.hooks[:]
		listeners := p.listeners[:]

		p.mu.RUnlock()

		for i := len(hooks) - 1; i >= 0; i-- {
			hooks[i].Open(proc)
		}

		for _, listener := range listeners {
			go listener.Accept(proc)
		}
	}

	return writer
}

// Close closes all writers and clears linked input ports, hooks, and listeners.
func (p *OutPort) Close() {
	p.mu.Lock()
	defer p.mu.Unlock()

	for _, writer := range p.writers {
		writer.Close()
	}

	p.writers = make(map[*process.Process]*packet.Writer)
	p.ins = nil
	p.hooks = nil
	p.listeners = nil
}
