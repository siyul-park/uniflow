package symbol

// LoadHook defines an interface for handling events when a symbol is loaded.
type LoadHook interface {
	// Load is called to handle the loading of a symbol and may return an error.
	Load(*Symbol) error
}

// LoadHooks is a slice of LoadHook interfaces, processed sequentially.
type LoadHooks []LoadHook

type loadHook struct {
	load func(*Symbol) error
}

var _ LoadHook = (LoadHooks)(nil)
var _ LoadHook = (*loadHook)(nil)

// LoadFunc creates a new LoadHook from the provided function.
func LoadFunc(load func(*Symbol) error) LoadHook {
	return &loadHook{load: load}
}

func (h LoadHooks) Load(sb *Symbol) error {
	for _, hook := range h {
		if err := hook.Load(sb); err != nil {
			return err
		}
	}
	return nil
}

func (h *loadHook) Load(sb *Symbol) error {
	return h.load(sb)
}
