package transaction

type RollbackHook interface {
	Rollback() error
}

type RollbackHookFunc func() error

var _ RollbackHook = RollbackHookFunc(func() error { return nil })

func (f RollbackHookFunc) Rollback() error {
	return f()
}
