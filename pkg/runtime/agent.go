package runtime

import (
	"sync"
	"time"

	"github.com/gofrs/uuid"

	"github.com/siyul-park/uniflow/pkg/packet"
	"github.com/siyul-park/uniflow/pkg/port"
	"github.com/siyul-park/uniflow/pkg/process"
	"github.com/siyul-park/uniflow/pkg/symbol"
)

// Agent manages symbols, processes, and hooks for ports and packets.
type Agent struct {
	symbols   map[uuid.UUID]*symbol.Symbol
	processes map[uuid.UUID]*process.Process
	frames    map[uuid.UUID][]*Frame
	inbounds  map[uuid.UUID]map[string]port.OpenHook
	outbounds map[uuid.UUID]map[string]port.OpenHook
	watchers  Watchers
	mu        sync.RWMutex
}

var _ symbol.LoadHook = (*Agent)(nil)

var _ symbol.UnloadHook = (*Agent)(nil)

// NewAgent initializes and returns a new Agent.
func NewAgent() *Agent {
	return &Agent{
		symbols:   make(map[uuid.UUID]*symbol.Symbol),
		processes: make(map[uuid.UUID]*process.Process),
		frames:    make(map[uuid.UUID][]*Frame),
		inbounds:  make(map[uuid.UUID]map[string]port.OpenHook),
		outbounds: make(map[uuid.UUID]map[string]port.OpenHook),
	}
}

// Watch registers a new watcher. Returns false if the watcher is already registered.
func (a *Agent) Watch(watcher Watcher) bool {
	a.mu.Lock()
	defer a.mu.Unlock()

	for _, w := range a.watchers {
		if w == watcher {
			return false
		}
	}
	a.watchers = append(a.watchers, watcher)
	return true
}

// Unwatch removes a watcher. Returns true if the watcher is successfully removed.
func (a *Agent) Unwatch(watcher Watcher) bool {
	a.mu.Lock()
	defer a.mu.Unlock()

	for i, w := range a.watchers {
		if w == watcher {
			a.watchers = append(a.watchers[:i], a.watchers[i+1:]...)
			return true
		}
	}
	return false
}

// Symbols returns a list of all registered symbols.
func (a *Agent) Symbols() []*symbol.Symbol {
	a.mu.RLock()
	defer a.mu.RUnlock()

	symbols := make([]*symbol.Symbol, 0, len(a.symbols))
	for _, sym := range a.symbols {
		symbols = append(symbols, sym)
	}
	return symbols
}

// Symbol returns a symbol by UUID.
func (a *Agent) Symbol(id uuid.UUID) *symbol.Symbol {
	a.mu.RLock()
	defer a.mu.RUnlock()

	return a.symbols[id]
}

// Processes returns a list of all registered processes.
func (a *Agent) Processes() []*process.Process {
	a.mu.RLock()
	defer a.mu.RUnlock()

	procs := make([]*process.Process, 0, len(a.processes))
	for _, proc := range a.processes {
		procs = append(procs, proc)
	}
	return procs
}

// Process returns a process by UUID.
func (a *Agent) Process(id uuid.UUID) *process.Process {
	a.mu.RLock()
	defer a.mu.RUnlock()

	return a.processes[id]
}

// Frames returns the frames associated with a specific process UUID.
func (a *Agent) Frames(id uuid.UUID) []*Frame {
	a.mu.RLock()
	defer a.mu.RUnlock()

	return append([]*Frame(nil), a.frames[id]...)
}

// Load registers a symbol and its associated hooks for inbound and outbound ports.
func (a *Agent) Load(sym *symbol.Symbol) error {
	a.mu.Lock()
	defer a.mu.Unlock()

	inbounds := make(map[string]port.OpenHook)
	outbounds := make(map[string]port.OpenHook)

	a.symbols[sym.ID()] = sym
	a.inbounds[sym.ID()] = inbounds
	a.outbounds[sym.ID()] = outbounds

	for name, in := range sym.Ins() {
		hook := port.OpenHookFunc(func(proc *process.Process) {
			a.accept(proc)

			inboundHook, outboundHook := a.hooks(proc, sym, in, nil)

			reader := in.Open(proc)
			reader.AddInboundHook(inboundHook)
			reader.AddOutboundHook(outboundHook)
		})

		in.AddOpenHook(hook)
		inbounds[name] = hook
	}

	for name, out := range sym.Outs() {
		hook := port.OpenHookFunc(func(proc *process.Process) {
			a.accept(proc)

			inboundHook, outboundHook := a.hooks(proc, sym, nil, out)

			writer := out.Open(proc)
			writer.AddInboundHook(inboundHook)
			writer.AddOutboundHook(outboundHook)
		})

		out.AddOpenHook(hook)
		outbounds[name] = hook
	}
	return nil
}

// Unload removes a symbol and its hooks.
func (a *Agent) Unload(sym *symbol.Symbol) error {
	a.mu.Lock()
	defer a.mu.Unlock()

	for name, hook := range a.inbounds[sym.ID()] {
		in := sym.In(name)
		in.RemoveOpenHook(hook)
	}
	for name, hook := range a.outbounds[sym.ID()] {
		out := sym.Out(name)
		out.RemoveOpenHook(hook)
	}

	delete(a.inbounds, sym.ID())
	delete(a.outbounds, sym.ID())
	delete(a.symbols, sym.ID())
	return nil
}

// Close clears all symbols, processes, and registered watchers.
func (a *Agent) Close() {
	a.mu.Lock()
	defer a.mu.Unlock()

	a.symbols = make(map[uuid.UUID]*symbol.Symbol)
	a.processes = make(map[uuid.UUID]*process.Process)
	a.frames = make(map[uuid.UUID][]*Frame)
	a.watchers = nil
}

// accept registers a process and notifies watchers.
func (a *Agent) accept(proc *process.Process) {
	a.mu.RLock()
	if _, ok := a.processes[proc.ID()]; ok {
		a.mu.RUnlock()
		return
	}
	a.mu.RUnlock()

	a.mu.Lock()

	if _, ok := a.processes[proc.ID()]; ok {
		a.mu.Unlock()
		return
	}

	a.processes[proc.ID()] = proc
	if _, ok := a.frames[proc.ID()]; !ok {
		a.frames[proc.ID()] = nil
	}

	watchers := a.watchers

	a.mu.Unlock()

	proc.AddExitHook(process.ExitFunc(func(err error) {
		a.mu.Lock()
		defer a.mu.Unlock()

		delete(a.processes, proc.ID())
		delete(a.frames, proc.ID())
	}))

	watchers.OnProcess(proc)
}

// hooks sets up hooks for a symbol's inbound and outbound ports.
func (a *Agent) hooks(proc *process.Process, sym *symbol.Symbol, in *port.InPort, out *port.OutPort) (packet.Hook, packet.Hook) {
	inboundHook := packet.HookFunc(func(pck *packet.Packet) {
		a.mu.Lock()

		var frame *Frame
		for _, f := range a.frames[proc.ID()] {
			if f.Symbol == sym && (f.InPort == in || f.OutPort == out) && f.InPck == nil {
				f.InPck = pck
				f.InTime = time.Now()
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
				InPck:   pck,
				InTime:  time.Now(),
			}
			a.frames[proc.ID()] = append(a.frames[proc.ID()], frame)
		}

		watchers := a.watchers

		a.mu.Unlock()

		watchers.OnFrame(frame)
	})

	outboundHook := packet.HookFunc(func(pck *packet.Packet) {
		a.mu.Lock()

		var frame *Frame
		for _, f := range a.frames[proc.ID()] {
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
			a.frames[proc.ID()] = append(a.frames[proc.ID()], frame)
		}

		watchers := a.watchers

		a.mu.Unlock()

		watchers.OnFrame(frame)
	})
	return inboundHook, outboundHook
}
