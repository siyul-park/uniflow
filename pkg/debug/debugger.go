package debug

import (
	"sync"
	"time"

	"github.com/gofrs/uuid"
	"github.com/siyul-park/uniflow/pkg/packet"
	"github.com/siyul-park/uniflow/pkg/port"
	"github.com/siyul-park/uniflow/pkg/process"
	"github.com/siyul-park/uniflow/pkg/symbol"
)

// Debugger manages symbols, processes, and their listeners.
type Debugger struct {
	symbols   map[uuid.UUID]*symbol.Symbol
	processes map[uuid.UUID]*process.Process
	frames    map[uuid.UUID][]*Frame
	inbounds  map[uuid.UUID]map[string]port.Hook
	outbounds map[uuid.UUID]map[string]port.Hook
	watchers  []Watcher
	mu        sync.RWMutex
}

var _ symbol.LoadHook = (*Debugger)(nil)
var _ symbol.UnloadHook = (*Debugger)(nil)

// NewDebugger creates and returns a new Debugger instance.
func NewDebugger() *Debugger {
	return &Debugger{
		symbols:   make(map[uuid.UUID]*symbol.Symbol),
		processes: make(map[uuid.UUID]*process.Process),
		frames:    make(map[uuid.UUID][]*Frame),
		inbounds:  make(map[uuid.UUID]map[string]port.Hook),
		outbounds: make(map[uuid.UUID]map[string]port.Hook),
	}
}

// AddWatcher adds a watcher to the debugger. Returns false if the watcher already exists.
func (d *Debugger) AddWatcher(watcher Watcher) bool {
	d.mu.Lock()
	defer d.mu.Unlock()

	for _, w := range d.watchers {
		if w == watcher {
			return false
		}
	}

	d.watchers = append(d.watchers, watcher)
	return true
}

// Symbols returns a list of all symbol IDs managed by the debugger.
func (d *Debugger) Symbols() []uuid.UUID {
	d.mu.RLock()
	defer d.mu.RUnlock()

	ids := make([]uuid.UUID, 0, len(d.symbols))
	for id := range d.symbols {
		ids = append(ids, id)
	}
	return ids
}

// Symbol retrieves a symbol by its ID.
func (d *Debugger) Symbol(id uuid.UUID) (*symbol.Symbol, bool) {
	d.mu.RLock()
	defer d.mu.RUnlock()

	sym, exists := d.symbols[id]
	return sym, exists
}

// Processes returns a list of all process IDs managed by the debugger.
func (d *Debugger) Processes() []uuid.UUID {
	d.mu.RLock()
	defer d.mu.RUnlock()

	ids := make([]uuid.UUID, 0, len(d.processes))
	for id := range d.processes {
		ids = append(ids, id)
	}
	return ids
}

// Process retrieves a process by its ID.
func (d *Debugger) Process(id uuid.UUID) (*process.Process, bool) {
	d.mu.RLock()
	defer d.mu.RUnlock()

	proc, exists := d.processes[id]
	return proc, exists
}

// Frames retrieves all frames associated with a specific process ID.
func (d *Debugger) Frames(id uuid.UUID) ([]*Frame, bool) {
	d.mu.RLock()
	defer d.mu.RUnlock()

	frames, exists := d.frames[id]
	return frames, exists
}

// Load adds a symbol and its associated listeners to the debugger.
func (d *Debugger) Load(sym *symbol.Symbol) error {
	d.mu.Lock()
	defer d.mu.Unlock()

	inbounds := make(map[string]port.Hook)
	outbounds := make(map[string]port.Hook)

	d.symbols[sym.ID()] = sym
	d.inbounds[sym.ID()] = inbounds
	d.outbounds[sym.ID()] = outbounds

	for _, name := range sym.Ins() {
		in := sym.In(name)
		hook := port.HookFunc(func(proc *process.Process) {
			d.accept(proc)

			inboundHook, outboundHook := d.hooks(proc, sym, in, nil)

			reader := in.Open(proc)
			reader.AddInboundHook(inboundHook)
			reader.AddOutboundHook(outboundHook)
		})

		in.AddHook(hook)
		inbounds[name] = hook
	}

	for _, name := range sym.Outs() {
		out := sym.Out(name)
		hook := port.HookFunc(func(proc *process.Process) {
			d.accept(proc)

			inboundHook, outboundHook := d.hooks(proc, sym, nil, out)

			writer := out.Open(proc)
			writer.AddInboundHook(inboundHook)
			writer.AddOutboundHook(outboundHook)
		})

		out.AddHook(hook)
		outbounds[name] = hook
	}

	return nil
}

// Unload removes a symbol and its associated listeners from the debugger.
func (d *Debugger) Unload(sym *symbol.Symbol) error {
	d.mu.Lock()
	defer d.mu.Unlock()

	for name, hook := range d.inbounds[sym.ID()] {
		in := sym.In(name)
		in.RemoveHook(hook)
	}
	for name, hook := range d.outbounds[sym.ID()] {
		out := sym.Out(name)
		out.RemoveHook(hook)
	}

	delete(d.inbounds, sym.ID())
	delete(d.outbounds, sym.ID())
	delete(d.symbols, sym.ID())

	return nil
}

func (d *Debugger) accept(proc *process.Process) {
	d.mu.Lock()
	defer d.mu.Unlock()

	if _, exists := d.processes[proc.ID()]; !exists {
		d.processes[proc.ID()] = proc

		proc.AddExitHook(process.ExitFunc(func(err error) {
			d.mu.Lock()
			defer d.mu.Unlock()

			delete(d.processes, proc.ID())
			delete(d.frames, proc.ID())
		}))

		watchers := d.watchers[:]

		d.mu.Unlock()

		for i := len(watchers) - 1; i >= 0; i-- {
			watcher := watchers[i]
			watcher.HandleProcess(proc)
		}

		d.mu.Lock()
	}

	if _, exists := d.frames[proc.ID()]; !exists {
		d.frames[proc.ID()] = nil
	}
}

func (d *Debugger) hooks(proc *process.Process, sym *symbol.Symbol, in *port.InPort, out *port.OutPort) (packet.Hook, packet.Hook) {
	inboundHook := packet.HookFunc(func(pck *packet.Packet) {
		d.mu.Lock()

		frame := &Frame{
			Process: proc,
			Symbol:  sym,
			InPort:  in,
			OutPort: out,
			InPck:   pck,
			InTime:  time.Now(),
		}
		d.frames[proc.ID()] = append(d.frames[proc.ID()], frame)

		watchers := d.watchers[:]

		d.mu.Unlock()

		for i := len(watchers) - 1; i >= 0; i-- {
			watcher := watchers[i]
			watcher.HandleFrame(frame)
		}
	})

	outboundHook := packet.HookFunc(func(pck *packet.Packet) {
		d.mu.Lock()

		var frame *Frame
		for _, f := range d.frames[proc.ID()] {
			if f.Symbol == sym && (f.InPort == in || f.OutPort == out) && f.OutPck == nil {
				f.OutPck = pck
				f.OutTime = time.Now()
				frame = f
				break
			}
		}
		if frame == nil {
			frame = &Frame{
				Process: proc,
				Symbol:  sym,
				InPort:  in,
				OutPort: out,
				OutPck:  pck,
				OutTime: time.Now(),
			}
			d.frames[proc.ID()] = append(d.frames[proc.ID()], frame)
		}

		watchers := d.watchers[:]

		d.mu.Unlock()

		for i := len(watchers) - 1; i >= 0; i-- {
			watcher := watchers[i]
			watcher.HandleFrame(frame)
		}
	})

	return inboundHook, outboundHook
}
