package symbol

import "github.com/siyul-park/uniflow/pkg/node"

type (
	// UnloadHook is a hook that is called node.Node is unloaded.
	UnloadHook interface {
		Unload(n node.Node)
	}

	UnloadHookFunc func(n node.Node)
)

var _ UnloadHook = UnloadHookFunc(func(n node.Node){})

func (f UnloadHookFunc) Unload(n node.Node) {
	f(n)
}
