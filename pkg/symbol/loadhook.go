package symbol

import "github.com/siyul-park/uniflow/pkg/node"

// LoadHook handles symbol load events.
type LoadHook interface {
	Load(*Symbol) error
}

// LoadHooks is a collection of LoadHook instances.
type LoadHooks []LoadHook

// LoadListener registers and invokes LoadHook instances.
type LoadListener interface {
	Load(LoadHook) error
}

type loadHook struct {
	fn func(*Symbol) error
}

var _ LoadHook = (*loadHook)(nil)
var _ LoadHook = (LoadHooks)(nil)

// LoadFunc wraps a function as a LoadHook.
func LoadFunc(fn func(*Symbol) error) LoadHook {
	return &loadHook{fn: fn}
}

// LoadListenerHook creates a LoadHook for nodes implementing LoadListener.
func LoadListenerHook(hook LoadHook) LoadHook {
	return LoadFunc(func(symbol *Symbol) error {
		if listener, ok := node.Unwrap(symbol).(LoadListener); ok {
			return listener.Load(hook)
		}
		return nil
	})
}

// Load executes all LoadHooks sequentially.
func (hooks LoadHooks) Load(symbol *Symbol) error {
	for _, hook := range hooks {
		if err := hook.Load(symbol); err != nil {
			return err
		}
	}
	return nil
}

// Load executes the associated function.
func (h *loadHook) Load(symbol *Symbol) error {
	return h.fn(symbol)
}
