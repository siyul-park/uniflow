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
	frame   *agent.Frame
	next    chan *agent.Frame
	done    chan *agent.Frame
	mu      sync.RWMutex
}

var _ agent.Watcher = (*Breakpoint)(nil)

// NewBreakpoint creates a new Breakpoint with optional configurations.
func NewBreakpoint(options ...func(*Breakpoint)) *Breakpoint {
	b := &Breakpoint{
		next: make(chan *agent.Frame),
		done: make(chan *agent.Frame),
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

	frame, ok := <-b.next

	b.mu.Lock()
	b.frame = frame
	b.mu.Unlock()
	return ok
}

// Done completes the current frame's processing.
func (b *Breakpoint) Done() bool {
	b.mu.Lock()
	frame := b.frame
	b.frame = nil
	b.mu.Unlock()

	if frame == nil {
		return false
	}

	b.done <- frame
	return true
}

// Frame returns the current frame under lock protection.
func (b *Breakpoint) Frame() *agent.Frame {
	b.mu.RLock()
	defer b.mu.RUnlock()

	return b.frame
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
		b.next <- frame
		<-b.done
	}
}

// OnProcess is a no-op but required by the Watcher interface.
func (b *Breakpoint) OnProcess(*process.Process) {}

// Close cleans up resources.
func (b *Breakpoint) Close() {
	b.mu.Lock()
	defer b.mu.Unlock()

	b.frame = nil

	close(b.next)
	close(b.done)
}

func (b *Breakpoint) matches(frame *agent.Frame) bool {
	return (b.process == nil || b.process == frame.Process) &&
		(b.symbol == nil || b.symbol == frame.Symbol) &&
		(b.inPort == nil || b.inPort == frame.InPort) &&
		(b.outPort == nil || b.outPort == frame.OutPort)
}
