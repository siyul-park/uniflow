package process

// CloseHook is an interface for hooks that are called when process.Process is closing.
type CloseHook interface {
	Close() error
}

// CloseHookFunc is a function type that implements the CloseHook interface.
type CloseHookFunc func() error

var _ CloseHook = CloseHookFunc(func() error { return nil })

func (f CloseHookFunc) Close() error {
	return f()
}
