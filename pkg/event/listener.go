package event

type Listener interface {
	Handel(e *Event) error
}

type ListenerFunc func(e *Event) error

var _ Listener = ListenerFunc(func(e *Event) error { return nil })

func (l ListenerFunc) Handel(e *Event) error {
	return l(e)
}
