package process

// ExitHook represents an interface for handling errors during process termination.
type ExitHook interface {
	Exit(err error)
}

type exitHook struct {
	exit func(err error)
}

var _ ExitHook = (*exitHook)(nil)

// ExitFunc creates a new ExitHook from the provided function.
func ExitFunc(exit func(err error)) ExitHook {
	return &exitHook{exit: exit}
}

func (h *exitHook) Exit(err error) {
	h.exit(err)
}
