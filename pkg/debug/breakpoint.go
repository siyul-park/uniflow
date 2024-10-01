package debug

import (
	"sync"

	"github.com/siyul-park/uniflow/pkg/agent"
	"github.com/siyul-park/uniflow/pkg/port"
	"github.com/siyul-park/uniflow/pkg/process"
	"github.com/siyul-park/uniflow/pkg/symbol"
)

// Breakpoint represents a synchronization point in a process where execution can be paused and resumed.
type Breakpoint struct {
	process *process.Process
	symbol  *symbol.Symbol
	inPort  *port.InPort
	outPort *port.OutPort
	current *agent.Frame
	in      chan *agent.Frame
	out     chan *agent.Frame
	done    chan struct{}
	rmu     sync.RWMutex
	wmu     sync.Mutex
}

var _ agent.Watcher = (*Breakpoint)(nil)

// NewBreakpoint creates a new Breakpoint with optional configurations.
func NewBreakpoint(options ...func(*Breakpoint)) *Breakpoint {
	b := &Breakpoint{
		in:   make(chan *agent.Frame),
		out:  make(chan *agent.Frame),
		done: make(chan struct{}),
	}
	for _, opt := range options {
		opt(b)
	}
	return b
}

// WithProcess sets the process associated with the breakpoint.
func WithProcess(proc *process.Process) func(*Breakpoint) {
	return func(b *Breakpoint) { b.process = proc }
}

// WithSymbol sets the symbol associated with the breakpoint.
func WithSymbol(sb *symbol.Symbol) func(*Breakpoint) {
	return func(b *Breakpoint) { b.symbol = sb }
}

// WithInPort sets the input port associated with the breakpoint.
func WithInPort(port *port.InPort) func(*Breakpoint) {
	return func(b *Breakpoint) { b.inPort = port }
}

// WithOutPort sets the output port associated with the breakpoint.
func WithOutPort(port *port.OutPort) func(*Breakpoint) {
	return func(b *Breakpoint) { b.outPort = port }
}

// Next advances to the next frame, returning false if closed.
func (b *Breakpoint) Next() bool {
	b.Done()

	b.rmu.Lock()
	defer b.rmu.Unlock()

	if b.current != nil {
		return false
	}

	select {
	case b.current = <-b.in:
		return true
	case <-b.done:
		return false
	}
}

// Done completes the current frame's processing.
func (b *Breakpoint) Done() bool {
	b.rmu.Lock()
	defer b.rmu.Unlock()

	if b.current == nil {
		return true
	}

	select {
	case b.out <- b.current:
		b.current = nil
		return true
	case <-b.done:
		return false
	}
}

// Frame returns the current frame under lock protection.
func (b *Breakpoint) Frame() *agent.Frame {
	b.rmu.RLock()
	defer b.rmu.RUnlock()

	return b.current
}

// Process returns the associated process.
func (b *Breakpoint) Process() *process.Process {
	return b.process
}

// Symbol returns the associated symbol.
func (b *Breakpoint) Symbol() *symbol.Symbol {
	return b.symbol
}

// InPort returns the associated input port.
func (b *Breakpoint) InPort() *port.InPort {
	return b.inPort
}

// OutPort returns the associated output port.
func (b *Breakpoint) OutPort() *port.OutPort {
	return b.outPort
}

// OnFrame processes an incoming frame and synchronizes it.
func (b *Breakpoint) OnFrame(frame *agent.Frame) {
	if b.matches(frame) {
		select {
		case b.in <- frame:
		case <-b.done:
		}

		select {
		case <-b.out:
		case <-b.done:
		}
	}
}

// OnProcess is a no-op but required by the Watcher interface.
func (b *Breakpoint) OnProcess(*process.Process) {}

// Close cleans up resources.
func (b *Breakpoint) Close() {
	b.wmu.Lock()
	defer b.wmu.Unlock()

	select {
	case <-b.done:
		return
	default:
	}

	close(b.done)

	b.rmu.Lock()
	defer b.rmu.Unlock()

	b.current = nil
}

func (b *Breakpoint) matches(frame *agent.Frame) bool {
	return (b.process == nil || b.process == frame.Process) &&
		(b.symbol == nil || b.symbol == frame.Symbol) &&
		(b.inPort == nil || b.inPort == frame.InPort) &&
		(b.outPort == nil || b.outPort == frame.OutPort)
}
