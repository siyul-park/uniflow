package port

import "github.com/siyul-park/uniflow/pkg/process"

// Hook defines an interface for processing packets associated with a process.
type Hook interface {
	// Open processes the given process.
	Open(*process.Process)
}

type hook struct {
	open func(*process.Process)
}

var _ Hook = (*hook)(nil)

// HookFunc creates a new Hook from the provided function.
func HookFunc(open func(*process.Process)) Hook {
	return &hook{open: open}
}

func (h *hook) Open(proc *process.Process) {
	h.open(proc)
}
