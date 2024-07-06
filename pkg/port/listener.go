package port

import "github.com/siyul-park/uniflow/pkg/process"

// Listener is an interface that defines a method to handle processes.
type Listener interface {
	Accept(proc *process.Process)
}

// ListenFunc is an adapter that allows using ordinary functions as Listener implementations.
type ListenFunc func(proc *process.Process)

var _ Listener = (ListenFunc)(nil)

// Accept calls the underlying function represented by ListenFunc.
func (lf ListenFunc) Accept(proc *process.Process) {
	lf(proc)
}
