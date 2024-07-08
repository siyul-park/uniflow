package process

// ExitHook represents an interface for handling errors during process termination.
type ExitHook interface {
	Exit(err error) // Exit is called to handle an error during termination.
}

// ExitFunc is an adapter that allows ordinary functions to implement the ExitHook interface.
type ExitFunc func(err error)

var _ ExitHook = (ExitFunc)(nil)

// Exit calls the underlying function to implement the ExitHook interface.
func (h ExitFunc) Exit(err error) {
	h(err)
}
