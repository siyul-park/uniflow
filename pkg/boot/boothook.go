package boot

import "context"

// BootHook defines a hook that runs during the bootstrap process.
type BootHook interface {
	Boot(context.Context) error
}

// BootHookFunc is a function type that implements the BootHook interface.
type BootHookFunc func(context.Context) error

var _ BootHook = BootHookFunc(nil)

// Boot calls the BootHookFunc.
func (f BootHookFunc) Boot(ctx context.Context) error {
	return f(ctx)
}
