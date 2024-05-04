package transaction

// CommitHook is an interface for committing transactions.
type CommitHook interface {
	Commit() error
}

// CommitHookFunc is a function type that implements the CommitHook interface.
type CommitHookFunc func() error

var _ CommitHook = CommitHookFunc(func() error { return nil })

// Commit calls the function itself to implement the CommitHook interface.
func (f CommitHookFunc) Commit() error {
	return f()
}
