package symbol

import "github.com/siyul-park/uniflow/pkg/node"

// UnloadHook handles symbol unload events.
type UnloadHook interface {
	Unload(*Symbol) error
}

// UnloadHooks is a collection of UnloadHook instances, processed in reverse order.
type UnloadHooks []UnloadHook

// UnloadListener registers and invokes UnloadHook instances.
type UnloadListener interface {
	Unload(UnloadHook) error
}

type unloadHook struct {
	fn func(*Symbol) error
}

var _ UnloadHook = (*unloadHook)(nil)
var _ UnloadHook = (UnloadHooks)(nil)

// UnloadFunc wraps a function as an UnloadHook.
func UnloadFunc(fn func(*Symbol) error) UnloadHook {
	return &unloadHook{fn: fn}
}

// UnloadListenerHook creates an UnloadHook for nodes implementing UnloadListener.
func UnloadListenerHook(hook UnloadHook) UnloadHook {
	return UnloadFunc(func(symbol *Symbol) error {
		if listener, ok := node.Unwrap(symbol).(UnloadListener); ok {
			return listener.Unload(hook)
		}
		return nil
	})
}

// Unload executes all UnloadHooks in reverse order.
func (hooks UnloadHooks) Unload(symbol *Symbol) error {
	for i := len(hooks) - 1; i >= 0; i-- {
		if err := hooks[i].Unload(symbol); err != nil {
			return err
		}
	}
	return nil
}

// Unload executes the associated function.
func (h *unloadHook) Unload(symbol *Symbol) error {
	return h.fn(symbol)
}
