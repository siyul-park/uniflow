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

var _ LoadHook = (LoadHooks)(nil)
var _ LoadHook = (*loadHook)(nil)

// LoadFunc wraps a function as a LoadHook.
func LoadFunc(fn func(*Symbol) error) LoadHook {
	return &loadHook{fn: fn}
}

// LoadListenerHook creates a LoadHook for nodes implementing LoadListener.
func LoadListenerHook(hook LoadHook) LoadHook {
	return LoadFunc(func(symbol *Symbol) error {
		var n node.Node = symbol
		for n != nil {
			if listener, ok := n.(LoadListener); ok {
				if err := listener.Load(hook); err != nil {
					return err
				}
			}
			n = node.Unwrap(n)
		}
		return nil
	})
}

// Load executes all LoadHooks sequentially.
func (h LoadHooks) Load(symbol *Symbol) error {
	for _, hook := range h {
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
