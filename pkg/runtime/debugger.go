package runtime

import (
	"context"
	"sync"

	"github.com/siyul-park/uniflow/pkg/process"
	"github.com/siyul-park/uniflow/pkg/symbol"
)

// Debugger manages breakpoints and the debugging process.
type Debugger struct {
	agent       *Agent
	breakpoints []*Breakpoint
	current     *Breakpoint
	in          chan *Breakpoint
	done        chan struct{}
	rmu         sync.RWMutex
	wmu         sync.RWMutex
}

// NewDebugger creates a new Debugger instance with the specified agent.
func NewDebugger(agent *Agent) *Debugger {
	return &Debugger{
		agent: agent,
		in:    make(chan *Breakpoint),
		done:  make(chan struct{}),
	}
}

// AddBreakpoint adds a breakpoint and starts monitoring it.
func (d *Debugger) AddBreakpoint(bp *Breakpoint) bool {
	d.wmu.Lock()
	defer d.wmu.Unlock()

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
	d.wmu.Lock()
	defer d.wmu.Unlock()

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
	d.wmu.RLock()
	defer d.wmu.RUnlock()

	return append([]*Breakpoint(nil), d.breakpoints...)
}

// Pause blocks until a breakpoint is hit or monitoring is done.
func (d *Debugger) Pause(ctx context.Context) bool {
	d.rmu.Lock()
	defer d.rmu.Unlock()

	if d.current != nil {
		return true
	}

	select {
	case d.current = <-d.in:
		return true
	case <-d.done:
		return false
	case <-ctx.Done():
		return false
	}
}

// Step continues execution until the next breakpoint is hit.
func (d *Debugger) Step(ctx context.Context) bool {
	d.rmu.Lock()
	defer d.rmu.Unlock()

	if d.current != nil {
		go d.next(d.current)
	}

	select {
	case d.current = <-d.in:
		return true
	case <-d.done:
		return false
	case <-ctx.Done():
		return false
	}
}

// Breakpoint returns the currently active breakpoint.
func (d *Debugger) Breakpoint() *Breakpoint {
	if d.rmu.TryRLock() {
		defer d.rmu.RUnlock()
		return d.current
	}
	return nil
}

// Frame returns the frame of the current breakpoint.
func (d *Debugger) Frame() *Frame {
	if d.rmu.TryRLock() {
		defer d.rmu.RUnlock()
		if d.current != nil {
			return d.current.Frame()
		}
	}
	return nil
}

// Process retrieves the process linked to the current breakpoint.
func (d *Debugger) Process() *process.Process {
	if d.rmu.TryRLock() {
		defer d.rmu.RUnlock()
		if d.current != nil {
			frame := d.current.Frame()
			if frame != nil {
				return frame.Process
			}
		}
	}
	return nil
}

// Symbol retrieves the symbol for the frame at the current breakpoint.
func (d *Debugger) Symbol() *symbol.Symbol {
	if d.rmu.TryRLock() {
		defer d.rmu.RUnlock()
		if d.current != nil {
			frame := d.current.Frame()
			if frame != nil {
				return frame.Symbol
			}
			return d.current.Symbol()
		}
	}
	return nil
}

// Close stops monitoring breakpoints and releases resources.
func (d *Debugger) Close() {
	d.wmu.Lock()
	defer d.wmu.Unlock()

	select {
	case <-d.done:
		return
	default:
	}

	close(d.done)

	for _, bp := range d.breakpoints {
		bp.Close()
	}
	d.breakpoints = nil

	d.rmu.Lock()
	defer d.rmu.Unlock()

	d.current = nil
}

func (d *Debugger) next(bp *Breakpoint) {
	if bp.Next() {
		select {
		case d.in <- bp:
		case <-d.done:
		}
	}
}
