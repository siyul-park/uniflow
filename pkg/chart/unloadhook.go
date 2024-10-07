package chart

// UnloadHook defines an interface for handling events when a symbol is unloaded.
type UnloadHook interface {
	// Unload is called when a symbol is unloaded and may return an error.
	Unload(*Chart) error
}

// UnloadHooks is a slice of UnloadHook interfaces, processed in reverse order.
type UnloadHooks []UnloadHook

type unloadHook struct {
	unload func(*Chart) error
}

var _ UnloadHook = (UnloadHooks)(nil)
var _ UnloadHook = (*unloadHook)(nil)

// UnloadFunc creates a new UnloadHook from the provided function.
func UnloadFunc(unload func(*Chart) error) UnloadHook {
	return &unloadHook{unload: unload}
}

func (h UnloadHooks) Unload(chrt *Chart) error {
	for i := len(h) - 1; i >= 0; i-- {
		hook := h[i]
		if err := hook.Unload(chrt); err != nil {
			return err
		}
	}
	return nil
}

func (h *unloadHook) Unload(chrt *Chart) error {
	return h.unload(chrt)
}
