package port

import "github.com/siyul-park/uniflow/pkg/process"

// Listener is an interface that defines a method to handle processes.
type Listener interface {
	// Accept is called to handle a process.
	Accept(proc *process.Process)
}

// ListenFunc is an adapter that allows using ordinary functions as implementations of Listener.
type ListenFunc func(proc *process.Process)

var _ Listener = (ListenFunc)(nil)

// Accept calls the underlying function represented by ListenFunc.
func (f ListenFunc) Accept(proc *process.Process) {
	f(proc)
}
