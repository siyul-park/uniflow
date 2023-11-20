package symbol

import "github.com/siyul-park/uniflow/pkg/node"

type (
	// PreLoadHook is a hook that is called before a node.Node is loaded.
	PreLoadHook interface {
		PreLoad(n node.Node) error
	}

	PreLoadHookFunc func(n node.Node) error
)

var _ PreLoadHook = PreLoadHookFunc(func(n node.Node) error { return nil })

func (f PreLoadHookFunc) PreLoad(n node.Node) error {
	return f(n)
}
