package process

// ExitHook is an interface that defines the method to handle errors during exit.
type ExitHook interface {
	Exit(err error)
}

// ExitFunc is an adapter that allows the use of ordinary functions as ExitHook implementations.
type ExitFunc func(err error)

var _ ExitHook = (ExitFunc)(nil)

// Exit calls the underlying function for the ExitFunc.
func (h ExitFunc) Exit(err error) {
	h(err)
}
