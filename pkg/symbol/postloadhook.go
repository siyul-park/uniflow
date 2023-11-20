package symbol

import "github.com/siyul-park/uniflow/pkg/node"

type (
	// PostLoadHook is a hook that is called after a node is loaded.
	PostLoadHook interface {
		PostLoad(n node.Node) error
	}

	PostLoadHookFunc func(n node.Node) error
)

var _ PostLoadHook = PostLoadHookFunc(func(n node.Node) error { return nil })

func (f PostLoadHookFunc) PostLoad(n node.Node) error {
	return f(n)
}
