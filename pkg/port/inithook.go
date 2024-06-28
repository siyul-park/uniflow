package port

import "github.com/siyul-park/uniflow/pkg/process"

// InitHook is an interface that defines the Serve method to handle processes.
type InitHook interface {
	Init(proc *process.Process)
}

// InitHookFunc is an adapter to allow the use of ordinary functions as Handlers.
type InitHookFunc func(proc *process.Process)

var _ InitHook = (InitHookFunc)(nil)

// Serve calls the underlying function for the HandlerFunc.
func (h InitHookFunc) Init(proc *process.Process) {
	h(proc)
}
