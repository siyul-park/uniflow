package port

import "github.com/siyul-park/uniflow/pkg/process"

// Hook defines an interface for processing packets associated with a process.
type Hook interface {
	// Open processes the given process.
	Open(*process.Process)
}

// Hooks is a slice of Hook interfaces, processed in reverse order.
type Hooks []Hook

type hook struct {
	open func(*process.Process)
}

var _ Hook = (Hooks)(nil)
var _ Hook = (*hook)(nil)

// HookFunc creates a new Hook from the provided function.
func HookFunc(open func(*process.Process)) Hook {
	return &hook{open: open}
}

func (h Hooks) Open(proc *process.Process) {
	for i := len(h) - 1; i >= 0; i-- {
		hook := h[i]
		hook.Open(proc)
	}
}

func (h *hook) Open(proc *process.Process) {
	h.open(proc)
}
