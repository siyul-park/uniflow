package port

// CloseHook is an interface that defines a method for handling resource cleanup.
type CloseHook interface {
	// Close performs the necessary cleanup actions.
	Close()
}

// CloseHooks represents a slice of CloseHook interfaces, which are processed in reverse order when closed.
type CloseHooks []CloseHook

type closeHook struct {
	close func()
}

var _ CloseHook = (CloseHooks)(nil)
var _ CloseHook = (*closeHook)(nil)

// CloseHookFunc creates a new CloseHook from the provided function.
func CloseHookFunc(fn func()) CloseHook {
	return &closeHook{close: fn}
}

func (h CloseHooks) Close() {
	for i := len(h) - 1; i >= 0; i-- {
		hook := h[i]
		hook.Close()
	}
}

func (h *closeHook) Close() {
	h.close()
}
