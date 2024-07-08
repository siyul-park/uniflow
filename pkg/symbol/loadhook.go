package symbol

// LoadHook is an interface for hooks that are called when a symbol is loaded.
type LoadHook interface {
	// Load is called when a symbol is loaded.
	Load(*Symbol) error
}

// LoadFunc is a function type that implements the LoadHook interface.
type LoadFunc func(*Symbol) error

var _ LoadHook = LoadFunc(nil)

// Load implements the Load method of the LoadHook interface.
func (f LoadFunc) Load(sym *Symbol) error {
	return f(sym)
}
