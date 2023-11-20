package symbol

import "github.com/siyul-park/uniflow/pkg/node"

type (
	// PreUnloadHook is a hook that is called before a node.Node is unloaded.
	PreUnloadHook interface {
		PreUnload(n node.Node) error
	}

	PreUnloadHookFunc func(n node.Node) error
)

var _ PreUnloadHook = PreUnloadHookFunc(func(n node.Node) error { return nil })

func (f PreUnloadHookFunc) PreUnload(n node.Node) error {
	return f(n)
}
