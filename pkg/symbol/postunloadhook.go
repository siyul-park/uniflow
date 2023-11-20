package symbol

import "github.com/siyul-park/uniflow/pkg/node"

type (
	// PostUnloadHook is a hook that is called after a node is unloaded.
	PostUnloadHook interface {
		PostUnload(n node.Node) error
	}

	PostUnloadHookFunc func(n node.Node) error
)

var _ PostUnloadHook = PostUnloadHookFunc(func(n node.Node) error { return nil })

func (f PostUnloadHookFunc) PostUnload(n node.Node) error {
	return f(n)
}
