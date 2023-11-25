package symbol

import "github.com/siyul-park/uniflow/pkg/node"

type (
	// LoadHook is a hook that is called node.Node is loaded.
	LoadHook interface {
		Load(n node.Node)
	}

	LoadHookFunc func(n node.Node)
)

var _ LoadHook = LoadHookFunc(func(n node.Node) { })

func (f LoadHookFunc) Load(n node.Node) {
	f(n)
}
