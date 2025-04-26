package process

// ExitHook represents an interface for handling errors during process termination.
type ExitHook interface {
	// Exit is called when the process terminates with an error.
	Exit(err error)
}

type exitHook struct {
	exit func(err error)
}

// ExitHooks is a slice of ExitHook interfaces, processed in reverse order.
type ExitHooks []ExitHook

var (
	_ ExitHook = (ExitHooks)(nil)
	_ ExitHook = (*exitHook)(nil)
)

// ExitFunc creates a new ExitHook from the provided function.
func ExitFunc(exit func(err error)) ExitHook {
	return &exitHook{exit: exit}
}

func (h ExitHooks) Exit(err error) {
	for i := len(h) - 1; i >= 0; i-- {
		hook := h[i]
		hook.Exit(err)
	}
}

func (h *exitHook) Exit(err error) {
	h.exit(err)
}
