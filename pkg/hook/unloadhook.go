package hook

import "github.com/siyul-park/uniflow/pkg/node"

// UnloadHook is an interface for hooks that are called when node.Node is unloaded.
type UnloadHook interface {
	// Unload is called when node.Node is unloaded.
	Unload(n node.Node) error
}

// UnloadHookFunc is a function type that implements the UnloadHook interface.
type UnloadHookFunc func(n node.Node) error

var _ UnloadHook = UnloadHookFunc(func(n node.Node) error { return nil })

// Unload is the implementation of the Unload method for UnloadHookFunc.
func (f UnloadHookFunc) Unload(n node.Node) error {
	return f(n)
}
