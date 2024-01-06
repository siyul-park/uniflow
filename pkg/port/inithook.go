package port

import (
	"sync"

	"github.com/siyul-park/uniflow/pkg/process"
)

// InitHook is a hook that is called when Port is initialized by Process.
type InitHook interface {
	Init(proc *process.Process)
}

type InitOnceHook struct {
	Hook  InitHook
	inits map[*process.Process]struct{}
	mu    sync.Mutex
}

type InitHookFunc func(proc *process.Process)

var _ InitHook = InitHookFunc(func(proc *process.Process) {})
var _ InitHook = (*InitOnceHook)(nil)

func (h InitHookFunc) Init(proc *process.Process) {
	h(proc)
}

func (h *InitOnceHook) Init(proc *process.Process) {
	if func() bool {
		h.mu.Lock()
		defer h.mu.Unlock()

		if h.inits == nil {
			h.inits = make(map[*process.Process]struct{})
		}
		
		if _, ok := h.inits[proc]; ok {
			return false
		}

		h.inits[proc] = struct{}{}
		go func() {
			<-proc.Done()

			h.mu.Lock()
			defer h.mu.Unlock()

			delete(h.inits, proc)
		}()

		return true
	}() {
		h.Hook.Init(proc)
	}
}
