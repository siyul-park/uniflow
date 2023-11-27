package symbol

import "github.com/siyul-park/uniflow/pkg/node"

type (
	// UnloadHook is a hook that is called node.Node is unloaded.
	UnloadHook interface {
		Unload(n node.Node) error
	}

	UnloadHookFunc func(n node.Node) error
)

var _ UnloadHook = UnloadHookFunc(func(n node.Node) error { return nil })

func (f UnloadHookFunc) Unload(n node.Node) error {
	return f(n)
}
