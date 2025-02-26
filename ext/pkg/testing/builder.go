package testing

import (
	"github.com/siyul-park/uniflow/pkg/hook"
	"github.com/siyul-park/uniflow/pkg/node"
	"github.com/siyul-park/uniflow/pkg/symbol"
	"github.com/siyul-park/uniflow/pkg/testing"
)

// AddToHook returns a function that adds hook to the provided hook.
func AddToHook(runner *testing.Runner) hook.Register {
	return hook.RegisterFunc(func(h *hook.Hook) error {
		h.AddLoadHook(symbol.LoadFunc(func(sb *symbol.Symbol) error {
			var n *TestNode
			if node.As(sb, &n) {
				runner.Register(sb.NamespacedName(), n)
			}
			return nil
		}))
		h.AddUnloadHook(symbol.UnloadFunc(func(sb *symbol.Symbol) error {
			var n *TestNode
			if node.As(sb, &n) {
				runner.Unregister(sb.NamespacedName())
			}
			return nil
		}))
		return nil
	})
}
