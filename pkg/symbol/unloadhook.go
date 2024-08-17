package symbol

// UnloadHook defines an interface for handling events when a symbol is unloaded.
type UnloadHook interface {
	// Unload is called when a symbol is unloaded.
	Unload(*Symbol) error
}

type unloadHook struct {
	unload func(*Symbol) error
}

var _ UnloadHook = (*unloadHook)(nil)

// UnloadFunc creates a new UnloadHook from the provided function.
func UnloadFunc(unload func(*Symbol) error) UnloadHook {
	return &unloadHook{unload: unload}
}

func (h *unloadHook) Unload(sym *Symbol) error {
	return h.unload(sym)
}
