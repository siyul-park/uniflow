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
	inbounds  map[uuid.UUID]map[string]port.OpenHook
	outbounds map[uuid.UUID]map[string]port.OpenHook
	watchers  Watchers
	mu        sync.RWMutex
}

var _ symbol.LoadHook = (*Agent)(nil)
var _ symbol.UnloadHook = (*Agent)(nil)

// New creates a new Agent instance.
func New() *Agent {
	return &Agent{
		symbols:   make(map[uuid.UUID]*symbol.Symbol),
		processes: make(map[uuid.UUID]*process.Process),
		frames:    make(map[uuid.UUID][]*Frame),
		inbounds:  make(map[uuid.UUID]map[string]port.OpenHook),
		outbounds: make(map[uuid.UUID]map[string]port.OpenHook),
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
func (a *Agent) Symbol(id uuid.UUID) *symbol.Symbol {
	a.mu.RLock()
	defer a.mu.RUnlock()

	return a.symbols[id]
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
func (a *Agent) Process(id uuid.UUID) *process.Process {
	a.mu.RLock()
	defer a.mu.RUnlock()

	return a.processes[id]
}

// Frames retrieves frames for a specific process UUID, returning frames and a boolean indicating existence.
func (a *Agent) Frames(id uuid.UUID) []*Frame {
	a.mu.RLock()
	defer a.mu.RUnlock()

	return append([]*Frame(nil), a.frames[id]...)
}

// Load adds a symbol and its hooks to the agent.
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

// Unload removes a symbol and its hooks from the agent.
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

	_, ok := a.processes[proc.ID()]
	if ok {
		a.mu.Unlock()
		return
	}

	a.processes[proc.ID()] = proc

	proc.AddExitHook(process.ExitFunc(func(err error) {
		a.mu.Lock()
		defer a.mu.Unlock()

		delete(a.processes, proc.ID())
		delete(a.frames, proc.ID())
	}))

	if _, ok := a.frames[proc.ID()]; !ok {
		a.frames[proc.ID()] = nil
	}

	watchers := a.watchers

	a.mu.Unlock()

	for i := len(watchers) - 1; i >= 0; i-- {
		watcher := watchers[i]
		watcher.OnProcess(proc)
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
