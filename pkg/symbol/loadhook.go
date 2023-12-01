package symbol

import "github.com/siyul-park/uniflow/pkg/node"

type (
	// LoadHook is an interface for hooks that are called when node.Node is loaded.
	LoadHook interface {
		// Load is called when node.Node is loaded.
		Load(n node.Node) error
	}

	// LoadHookFunc is a function type that implements the LoadHook interface.
	LoadHookFunc func(n node.Node) error
)

var _ LoadHook = LoadHookFunc(func(n node.Node) error { return nil })

// Load is the implementation of the Load method for LoadHookFunc.
func (f LoadHookFunc) Load(n node.Node) error {
	return f(n)
}
