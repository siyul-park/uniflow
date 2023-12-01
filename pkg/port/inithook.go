package port

import (
	"github.com/siyul-park/uniflow/pkg/process"
)

// InitHook is a hook that is called when Port is initialized by Process.
type InitHook interface {
	Init(proc *process.Process)
}

// InitHookFunc is a function type that implements the InitHook interface.
type InitHookFunc func(proc *process.Process)

// Ensure InitHookFunc implements the InitHook interface.
var _ InitHook = InitHookFunc(func(proc *process.Process) {})

// Init calls the underlying function for InitHookFunc.
func (h InitHookFunc) Init(proc *process.Process) {
	h(proc)
}
