package port

import "github.com/siyul-park/uniflow/pkg/process"

// OpenHook defines an interface for processing packets associated with a process.
type OpenHook interface {
	// Open processes the given process.
	Open(*process.Process)
}

// OpenHooks is a slice of Hook interfaces, processed in reverse order.
type OpenHooks []OpenHook

type openHook struct {
	open func(*process.Process)
}

var _ OpenHook = (OpenHooks)(nil)

var _ OpenHook = (*openHook)(nil)

// OpenHookFunc creates a new Hook from the provided function.
func OpenHookFunc(open func(*process.Process)) OpenHook {
	return &openHook{open: open}
}

func (h OpenHooks) Open(proc *process.Process) {
	for i := len(h) - 1; i >= 0; i-- {
		hook := h[i]
		hook.Open(proc)
	}
}

func (h *openHook) Open(proc *process.Process) {
	h.open(proc)
}
