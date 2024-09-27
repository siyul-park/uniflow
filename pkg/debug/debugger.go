package debug

import (
	"context"
	"sync"

	"github.com/siyul-park/uniflow/pkg/agent"
	"github.com/siyul-park/uniflow/pkg/process"
	"github.com/siyul-park/uniflow/pkg/symbol"
)

// Debugger manages breakpoints and the debugging process.
type Debugger struct {
	agent       *agent.Agent
	breakpoints []*Breakpoint
	current     *Breakpoint
	nexts       chan *Breakpoint
	done        chan struct{}
	mu          sync.RWMutex
}

// NewDebugger creates a new Debugger instance with the specified agent.
func NewDebugger(agent *agent.Agent) *Debugger {
	return &Debugger{
		agent: agent,
		nexts: make(chan *Breakpoint),
		done:  make(chan struct{}),
	}
}

// AddBreakpoint adds a breakpoint and starts monitoring it.
func (d *Debugger) AddBreakpoint(bp *Breakpoint) bool {
	d.mu.Lock()
	defer d.mu.Unlock()

	for _, b := range d.breakpoints {
		if b == bp {
			return false
		}
	}

	d.breakpoints = append(d.breakpoints, bp)
	d.agent.Watch(bp)

	go d.next(bp)
	return true
}

// RemoveBreakpoint deletes the specified breakpoint.
func (d *Debugger) RemoveBreakpoint(bp *Breakpoint) bool {
	d.mu.Lock()
	defer d.mu.Unlock()

	for i, b := range d.breakpoints {
		if b == bp {
			d.breakpoints = append(d.breakpoints[:i], d.breakpoints[i+1:]...)
			d.agent.Unwatch(bp)

			bp.Close()
			return true
		}
	}
	return false
}

// Breakpoints returns all registered breakpoints.
func (d *Debugger) Breakpoints() []*Breakpoint {
	d.mu.RLock()
	defer d.mu.RUnlock()

	if len(d.breakpoints) == 0 {
		return nil
	}
	return append([]*Breakpoint(nil), d.breakpoints...)
}

// Pause blocks until a breakpoint is hit or monitoring is done.
func (d *Debugger) Pause(ctx context.Context) bool {
	d.mu.RLock()
	defer d.mu.RUnlock()

	if d.current != nil {
		return true
	}

	select {
	case d.current = <-d.nexts:
		return true
	case <-d.done:
		return false
	case <-ctx.Done():
		return false
	}
}

// Step continues execution until the next breakpoint is hit.
func (d *Debugger) Step(ctx context.Context) bool {
	d.mu.RLock()
	defer d.mu.RUnlock()

	if d.current != nil {
		go d.next(d.current)
	}

	select {
	case d.current = <-d.nexts:
		return true
	case <-d.done:
		return false
	case <-ctx.Done():
		return false
	}
}

// Breakpoint returns the currently active breakpoint.
func (d *Debugger) Breakpoint() *Breakpoint {
	d.mu.RLock()
	defer d.mu.RUnlock()

	return d.current
}

// Frame returns the frame of the current breakpoint.
func (d *Debugger) Frame() *agent.Frame {
	d.mu.RLock()
	defer d.mu.RUnlock()

	if d.current != nil {
		return d.current.Frame()
	}
	return nil
}

// Process retrieves the process linked to the current breakpoint.
func (d *Debugger) Process() *process.Process {
	d.mu.RLock()
	defer d.mu.RUnlock()

	if d.current != nil {
		frame := d.current.Frame()
		if frame != nil {
			return frame.Process
		}
	}
	return nil
}

// Symbol retrieves the symbol for the frame at the current breakpoint.
func (d *Debugger) Symbol() *symbol.Symbol {
	d.mu.RLock()
	defer d.mu.RUnlock()

	if d.current != nil {
		frame := d.current.Frame()
		if frame != nil {
			return frame.Symbol
		}
		return d.current.Symbol()
	}
	return nil
}

// Close stops monitoring breakpoints and releases resources.
func (d *Debugger) Close() {
	d.mu.Lock()
	defer d.mu.Unlock()

	select {
	case <-d.done:
		return
	default:
		close(d.done)

		for _, bp := range d.breakpoints {
			bp.Close()
		}
		d.breakpoints = nil

		close(d.nexts)
	}
}

func (d *Debugger) next(bp *Breakpoint) {
	if bp.Next() {
		select {
		case d.nexts <- bp:
		case <-d.done:
		}
	}
}
