package debug

import (
	"sync"

	"github.com/gofrs/uuid"
	"github.com/siyul-park/uniflow/pkg/port"
	"github.com/siyul-park/uniflow/pkg/process"
	"github.com/siyul-park/uniflow/pkg/symbol"
)

// Debugger manages symbols, processes, and their listeners.
type Debugger struct {
	symbols   map[uuid.UUID]*symbol.Symbol
	processes map[uuid.UUID]*process.Process
	inbounds  map[uuid.UUID]map[string]port.Listener
	outbounds map[uuid.UUID]map[string]port.Listener
	mu        sync.RWMutex
}

var _ symbol.LoadHook = (*Debugger)(nil)
var _ symbol.UnloadHook = (*Debugger)(nil)

// NewDebugger creates a new Debugger instance.
func NewDebugger() *Debugger {
	return &Debugger{
		symbols:   make(map[uuid.UUID]*symbol.Symbol),
		processes: make(map[uuid.UUID]*process.Process),
		inbounds:  make(map[uuid.UUID]map[string]port.Listener),
		outbounds: make(map[uuid.UUID]map[string]port.Listener),
	}
}

// Symbol retrieves a symbol by its ID.
func (d *Debugger) Symbol(id uuid.UUID) (*symbol.Symbol, bool) {
	d.mu.RLock()
	defer d.mu.RUnlock()

	sym, exists := d.symbols[id]
	return sym, exists
}

// Process retrieves a process by its ID.
func (d *Debugger) Process(id uuid.UUID) (*process.Process, bool) {
	d.mu.RLock()
	defer d.mu.RUnlock()

	proc, exists := d.processes[id]
	return proc, exists
}

// Load adds a symbol to the debugger and sets up listeners for its ports.
func (d *Debugger) Load(sym *symbol.Symbol) error {
	d.mu.Lock()
	defer d.mu.Unlock()

	inbounds := make(map[string]port.Listener)
	outbounds := make(map[string]port.Listener)

	d.symbols[sym.ID()] = sym
	d.inbounds[sym.ID()] = inbounds
	d.outbounds[sym.ID()] = outbounds

	for _, name := range sym.Ins() {
		in := sym.In(name)
		listener := port.ListenFunc(d.accept)

		in.AddListener(listener)
		inbounds[name] = listener
	}

	for _, name := range sym.Outs() {
		out := sym.Out(name)
		listener := port.ListenFunc(d.accept)

		out.AddListener(listener)
		outbounds[name] = listener
	}

	return nil
}

// Unload removes a symbol and its associated listeners from the debugger.
func (d *Debugger) Unload(sym *symbol.Symbol) error {
	d.mu.Lock()
	defer d.mu.Unlock()

	for name, listener := range d.inbounds[sym.ID()] {
		in := sym.In(name)
		in.RemoveListener(listener)
	}
	for name, listener := range d.outbounds[sym.ID()] {
		out := sym.Out(name)
		out.RemoveListener(listener)
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
		}))
	}
}
