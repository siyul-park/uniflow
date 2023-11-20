package port

import (
	"sync"

	"github.com/siyul-park/uniflow/pkg/process"
)

type (
	// InitHook is a hook that is called when Port is initialized by Process.
	InitHook interface {
		Init(proc *process.Process)
	}

	InitHookFunc func(proc *process.Process)

	// InitOnceHook is a hook that runs only once per process.process.
	InitOnceHook struct {
		init      InitHook
		processes map[*process.Process]struct{}
		mu        sync.RWMutex
	}
)

var _ InitHook = InitHookFunc(func(proc *process.Process) {})
var _ InitHook = &InitOnceHook{}

func (h InitHookFunc) Init(proc *process.Process) {
	h(proc)
}

// InitOnce returns a new InitOnceHook.
func InitOnce(h InitHook) *InitOnceHook {
	return &InitOnceHook{
		init:      h,
		processes: make(map[*process.Process]struct{}),
	}
}

func (h *InitOnceHook) Init(proc *process.Process) {
	if ok := func() bool {
		h.mu.RLock()
		defer h.mu.RUnlock()

		_, ok := h.processes[proc]
		return !ok
	}(); !ok {
		return
	}

	if ok := func() bool {
		h.mu.Lock()
		defer h.mu.Unlock()

		_, ok := h.processes[proc]
		if ok {
			return false
		}

		h.processes[proc] = struct{}{}
		go func() {
			<-proc.Done()

			h.mu.Lock()
			defer h.mu.Unlock()

			delete(h.processes, proc)
		}()

		return true
	}(); !ok {
		return
	}

	h.init.Init(proc)
}

func (h *InitOnceHook) Close() {
	h.mu.Lock()
	defer h.mu.Unlock()

	for proc := range h.processes {
		delete(h.processes, proc)
	}
}
