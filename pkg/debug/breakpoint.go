package debug

import (
	"sync"

	"github.com/siyul-park/uniflow/pkg/port"
	"github.com/siyul-park/uniflow/pkg/process"
	"github.com/siyul-park/uniflow/pkg/symbol"
)

// BreakPoint represents a synchronization point in a process where execution can be paused and resumed.
type BreakPoint struct {
	process *process.Process
	symbol  *symbol.Symbol
	inPort  *port.InPort
	outPort *port.OutPort
	frame   *Frame
	next    chan *Frame
	done    chan *Frame
	mu      sync.RWMutex
}

var _ Watcher = (*BreakPoint)(nil)

// NewBreakPoint creates a new BreakPoint with optional configurations.
func NewBreakPoint(options ...func(*BreakPoint)) *BreakPoint {
	b := &BreakPoint{
		next: make(chan *Frame),
		done: make(chan *Frame),
	}
	for _, opt := range options {
		opt(b)
	}
	return b
}

// WithProcess configures the process associated with the breakpoint.
func WithProcess(proc *process.Process) func(*BreakPoint) {
	return func(b *BreakPoint) {
		b.process = proc
	}
}

// WithSymbol configures the symbol associated with the breakpoint.
func WithSymbol(sym *symbol.Symbol) func(*BreakPoint) {
	return func(b *BreakPoint) {
		b.symbol = sym
	}
}

// WithInPort configures the input port associated with the breakpoint.
func WithInPort(port *port.InPort) func(*BreakPoint) {
	return func(b *BreakPoint) {
		b.inPort = port
	}
}

// WithOutPort configures the output port associated with the breakpoint.
func WithOutPort(port *port.OutPort) func(*BreakPoint) {
	return func(b *BreakPoint) {
		b.outPort = port
	}
}

// Next advances to the next frame and returns false if the channel is closed.
func (b *BreakPoint) Next() bool {
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

// Frame returns the current frame under lock protection.
func (b *BreakPoint) Frame() *Frame {
	b.mu.RLock()
	defer b.mu.RUnlock()

	return b.frame
}

// HandleFrame processes an incoming frame and synchronizes it.
func (b *BreakPoint) HandleFrame(frame *Frame) {
	if b.watch(frame) {
		b.next <- frame
		<-b.done
	}
}

// HandleProcess is currently a no-op but required by the Watcher interface.
func (b *BreakPoint) HandleProcess(*process.Process) {
	// No operation; method required to satisfy the Watcher interface.
}

// Close closes the next channel to signal the end of the BreakPoint's lifecycle.
func (b *BreakPoint) Close() {
	close(b.next)
}

func (b *BreakPoint) watch(frame *Frame) bool {
	return (b.process == nil || b.process == frame.Process) &&
		(b.symbol == nil || b.symbol == frame.Symbol) &&
		(b.inPort == nil || b.inPort == frame.InPort) &&
		(b.outPort == nil || b.outPort == frame.OutPort)
}
