package symbol

// UnloadHook is an interface for hooks that are called when node.Node is unloaded.
type UnloadHook interface {
	// Unload is called when node.Node is unloaded.
	Unload(*Symbol) error
}

// UnloadHookFunc is a function type that implements the UnloadHook interface.
type UnloadHookFunc func(*Symbol) error

var _ UnloadHook = UnloadHookFunc(nil)

// Unload is the implementation of the Unload method for UnloadHookFunc.
func (f UnloadHookFunc) Unload(sym *Symbol) error {
	return f(sym)
}
