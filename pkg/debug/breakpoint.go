package debug

import (
	"sync"

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
	frame   *Frame
	next    chan *Frame
	done    chan *Frame
	mu      sync.RWMutex
}

var _ Watcher = (*Breakpoint)(nil)

// NewBreakpoint creates a new Breakpoint with optional configurations.
func NewBreakpoint(options ...func(*Breakpoint)) *Breakpoint {
	b := &Breakpoint{
		next: make(chan *Frame),
		done: make(chan *Frame),
	}
	for _, opt := range options {
		opt(b)
	}
	return b
}

// WithProcess sets the process associated with the breakpoint.
func WithProcess(proc *process.Process) func(*Breakpoint) {
	return func(b *Breakpoint) {
		b.process = proc
	}
}

// WithSymbol sets the symbol associated with the breakpoint.
func WithSymbol(sym *symbol.Symbol) func(*Breakpoint) {
	return func(b *Breakpoint) {
		b.symbol = sym
	}
}

// WithInPort sets the input port associated with the breakpoint.
func WithInPort(port *port.InPort) func(*Breakpoint) {
	return func(b *Breakpoint) {
		b.inPort = port
	}
}

// WithOutPort sets the output port associated with the breakpoint.
func WithOutPort(port *port.OutPort) func(*Breakpoint) {
	return func(b *Breakpoint) {
		b.outPort = port
	}
}

// Next advances to the next frame and returns false if the channel is closed.
func (b *Breakpoint) Next() bool {
	b.mu.Lock()
	defer b.mu.Unlock()

	if b.frame != nil {
		b.done <- b.frame
	}

	frame, ok := <-b.next
	if !ok {
		b.frame = nil
		close(b.done)
		return false
	}

	b.frame = frame
	return true
}

// Done completes the current frame's processing.
func (b *Breakpoint) Done() bool {
	b.mu.Lock()
	defer b.mu.Unlock()

	if b.frame == nil {
		return false
	}

	b.done <- b.frame
	b.frame = nil
	return true
}

// Frame returns the current frame under lock protection.
func (b *Breakpoint) Frame() *Frame {
	b.mu.RLock()
	defer b.mu.RUnlock()

	return b.frame
}

// Process returns the process associated with the breakpoint.
func (b *Breakpoint) Process() *process.Process {
	b.mu.RLock()
	defer b.mu.RUnlock()

	return b.process
}

// Symbol returns the symbol associated with the breakpoint.
func (b *Breakpoint) Symbol() *symbol.Symbol {
	b.mu.RLock()
	defer b.mu.RUnlock()

	return b.symbol
}

// InPort returns the input port associated with the breakpoint.
func (b *Breakpoint) InPort() *port.InPort {
	b.mu.RLock()
	defer b.mu.RUnlock()

	return b.inPort
}

// OutPort returns the output port associated with the breakpoint.
func (b *Breakpoint) OutPort() *port.OutPort {
	b.mu.RLock()
	defer b.mu.RUnlock()

	return b.outPort
}

// HandleFrame processes an incoming frame and synchronizes it with the breakpoint's criteria.
func (b *Breakpoint) HandleFrame(frame *Frame) {
	if b.watch(frame) {
		b.next <- frame
		<-b.done
	}
}

// HandleProcess is a no-op but required by the Watcher interface.
func (b *Breakpoint) HandleProcess(*process.Process) {
	// No operation; required for Watcher interface compliance.
}

// Close closes the next channel and cleans up resources.
func (b *Breakpoint) Close() {
	close(b.next)

	b.mu.Lock()
	defer b.mu.Unlock()

	if b.frame != nil {
		b.done <- b.frame
	}
	b.frame = nil
}

// watch checks if a frame matches the criteria of the breakpoint.
func (b *Breakpoint) watch(frame *Frame) bool {
	return (b.process == nil || b.process == frame.Process) &&
		(b.symbol == nil || b.symbol == frame.Symbol) &&
		(b.inPort == nil || b.inPort == frame.InPort) &&
		(b.outPort == nil || b.outPort == frame.OutPort)
}
