package symbol

// LoadHook is an interface for hooks that are called when node.Node is loaded.
type LoadHook interface {
	// Load is called when node.Node is loaded.
	Load(*Symbol) error
}

// LoadFunc is a function type that implements the LoadHook interface.
type LoadFunc func(*Symbol) error

var _ LoadHook = LoadFunc(nil)

// Load is the implementation of the Load method for LoadHookFunc.
func (f LoadFunc) Load(sym *Symbol) error {
	return f(sym)
}
