package symbol

import "github.com/siyul-park/uniflow/pkg/node"

type (
	// LoadHook is a hook that is called node.Node is loaded.
	LoadHook interface {
		Load(n node.Node) error
	}

	LoadHookFunc func(n node.Node) error
)

var _ LoadHook = LoadHookFunc(func(n node.Node) error { return nil })

func (f LoadHookFunc) Load(n node.Node) error {
	return f(n)
}
