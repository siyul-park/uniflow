package transaction

// RollbackHook is an interface for rolling back transactions.
type RollbackHook interface {
	Rollback() error
}

// RollbackHookFunc is a function type that implements the RollbackHook interface.
type RollbackHookFunc func() error

var _ RollbackHook = RollbackHookFunc(func() error { return nil })

// Rollback calls the function itself to implement the RollbackHook interface.
func (f RollbackHookFunc) Rollback() error {
	return f()
}
