package process

type ExitHook interface {
	Exit(err error)
}

type ExitHookFunc func(err error)

var _ ExitHook = (ExitHookFunc)(nil)

func (h ExitHookFunc) Exit(err error) {
	h(err)
}
