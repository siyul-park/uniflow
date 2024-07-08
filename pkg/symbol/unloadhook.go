package symbol

// UnloadHook is an interface for hooks that are called when a symbol is unloaded.
type UnloadHook interface {
	// Unload is called when a symbol is unloaded.
	Unload(*Symbol) error
}

// UnloadFunc is a function type that implements the UnloadHook interface.
type UnloadFunc func(*Symbol) error

var _ UnloadHook = UnloadFunc(nil)

// Unload implements the Unload method of the UnloadHook interface.
func (f UnloadFunc) Unload(sym *Symbol) error {
	return f(sym)
}
