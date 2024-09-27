package agent

import (
	"sync"
	"time"

	"github.com/gofrs/uuid"
	"github.com/siyul-park/uniflow/pkg/packet"
	"github.com/siyul-park/uniflow/pkg/port"
	"github.com/siyul-park/uniflow/pkg/process"
	"github.com/siyul-park/uniflow/pkg/symbol"
)

// Agent manages symbols, processes, and their hooks.
type Agent struct {
	symbols   map[uuid.UUID]*symbol.Symbol
	processes map[uuid.UUID]*process.Process
	frames    map[uuid.UUID][]*Frame
	inbounds  map[uuid.UUID]map[string]port.Hook
	outbounds map[uuid.UUID]map[string]port.Hook
	watchers  []Watcher
	mu        sync.RWMutex
}

// Ensure Agent implements LoadHook and UnloadHook interfaces.
var _ symbol.LoadHook = (*Agent)(nil)
var _ symbol.UnloadHook = (*Agent)(nil)

// New creates a new Agent instance.
func New() *Agent {
	return &Agent{
		symbols:   make(map[uuid.UUID]*symbol.Symbol),
		processes: make(map[uuid.UUID]*process.Process),
		frames:    make(map[uuid.UUID][]*Frame),
		inbounds:  make(map[uuid.UUID]map[string]port.Hook),
		outbounds: make(map[uuid.UUID]map[string]port.Hook),
	}
}

// Watch registers a watcher, returning false if already registered.
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

// Unwatch unregisters a watcher, returning true if successfully removed.
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

// Symbols returns all managed symbols.
func (a *Agent) Symbols() []*symbol.Symbol {
	a.mu.RLock()
	defer a.mu.RUnlock()

	sbs := make([]*symbol.Symbol, 0, len(a.symbols))
	for _, sb := range a.symbols {
		sbs = append(sbs, sb)
	}
	return sbs
}

// Symbol retrieves a symbol by UUID, returning the symbol and a boolean indicating existence.
func (a *Agent) Symbol(id uuid.UUID) (*symbol.Symbol, bool) {
	a.mu.RLock()
	defer a.mu.RUnlock()

	sym, exists := a.symbols[id]
	return sym, exists
}

// Processes returns all managed processes.
func (a *Agent) Processes() []*process.Process {
	a.mu.RLock()
	defer a.mu.RUnlock()

	procs := make([]*process.Process, 0, len(a.processes))
	for _, proc := range a.processes {
		procs = append(procs, proc)
	}
	return procs
}

// Process retrieves a process by UUID, returning the process and a boolean indicating existence.
func (a *Agent) Process(id uuid.UUID) (*process.Process, bool) {
	a.mu.RLock()
	defer a.mu.RUnlock()

	proc, exists := a.processes[id]
	return proc, exists
}

// Frames retrieves frames for a specific process UUID, returning frames and a boolean indicating existence.
func (a *Agent) Frames(id uuid.UUID) ([]*Frame, bool) {
	a.mu.RLock()
	defer a.mu.RUnlock()

	frames, exists := a.frames[id]
	if !exists {
		return nil, false
	}
	return append([]*Frame(nil), frames...), true
}

// Load adds a symbol and its hooks to the agent.
func (a *Agent) Load(sym *symbol.Symbol) error {
	a.mu.Lock()
	defer a.mu.Unlock()

	inbounds := make(map[string]port.Hook)
	outbounds := make(map[string]port.Hook)

	a.symbols[sym.ID()] = sym
	a.inbounds[sym.ID()] = inbounds
	a.outbounds[sym.ID()] = outbounds

	for _, name := range sym.Ins() {
		in := sym.In(name)
		hook := port.HookFunc(func(proc *process.Process) {
			a.accept(proc)

			inboundHook, outboundHook := a.hooks(proc, sym, in, nil)

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
			a.accept(proc)

			inboundHook, outboundHook := a.hooks(proc, sym, nil, out)

			writer := out.Open(proc)
			writer.AddInboundHook(inboundHook)
			writer.AddOutboundHook(outboundHook)
		})

		out.AddHook(hook)
		outbounds[name] = hook
	}

	return nil
}

// Unload removes a symbol and its hooks from the agent.
func (a *Agent) Unload(sym *symbol.Symbol) error {
	a.mu.Lock()
	defer a.mu.Unlock()

	for name, hook := range a.inbounds[sym.ID()] {
		in := sym.In(name)
		in.RemoveHook(hook)
	}
	for name, hook := range a.outbounds[sym.ID()] {
		out := sym.Out(name)
		out.RemoveHook(hook)
	}

	delete(a.inbounds, sym.ID())
	delete(a.outbounds, sym.ID())
	delete(a.symbols, sym.ID())

	return nil
}

// Close releases all resources and clears symbols, processes, and watchers.
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
	a.mu.Lock()
	defer a.mu.Unlock()

	if _, exists := a.processes[proc.ID()]; !exists {
		a.processes[proc.ID()] = proc

		proc.AddExitHook(process.ExitFunc(func(err error) {
			a.mu.Lock()
			defer a.mu.Unlock()

			delete(a.processes, proc.ID())
			delete(a.frames, proc.ID())
		}))

		for _, watcher := range a.watchers {
			watcher.OnProcess(proc)
		}

		if _, exists := a.frames[proc.ID()]; !exists {
			a.frames[proc.ID()] = nil
		}
	}
}

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

		watchers := a.watchers[:]
		a.mu.Unlock()

		for i := len(watchers) - 1; i >= 0; i-- {
			watcher := watchers[i]
			watcher.OnFrame(frame)
		}
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

		watchers := a.watchers[:]
		a.mu.Unlock()

		for i := len(watchers) - 1; i >= 0; i-- {
			watcher := watchers[i]
			watcher.OnFrame(frame)
		}
	})

	return inboundHook, outboundHook
}
