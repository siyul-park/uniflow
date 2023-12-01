package port

import (
	"github.com/siyul-park/uniflow/pkg/process"
)

type (
	// InitHook is a hook that is called when Port is initialized by Process.
	InitHook interface {
		Init(proc *process.Process)
	}

	InitHookFunc func(proc *process.Process)
)

var _ InitHook = InitHookFunc(func(proc *process.Process) {})

func (h InitHookFunc) Init(proc *process.Process) {
	h(proc)
}
