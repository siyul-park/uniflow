package transaction

type CommitHook interface {
	Commit() error
}

type CommitHookFunc func() error

var _ CommitHook = CommitHookFunc(func() error { return nil })

func (f CommitHookFunc) Commit() error {
	return f()
}
