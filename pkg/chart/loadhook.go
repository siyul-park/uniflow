package chart

// LoadHook defines an interface for handling events when a symbol is loaded.
type LoadHook interface {
	// Load is called to handle the loading of a symbol and may return an error.
	Load(*Chart) error
}

// LoadHooks is a slice of LoadHook interfaces, processed sequentially.
type LoadHooks []LoadHook

type loadHook struct {
	load func(*Chart) error
}

var _ LoadHook = (LoadHooks)(nil)
var _ LoadHook = (*loadHook)(nil)

// LoadFunc creates a new LoadHook from the provided function.
func LoadFunc(load func(*Chart) error) LoadHook {
	return &loadHook{load: load}
}

func (h LoadHooks) Load(chrt *Chart) error {
	for _, hook := range h {
		if err := hook.Load(chrt); err != nil {
			return err
		}
	}
	return nil
}

func (h *loadHook) Load(chrt *Chart) error {
	return h.load(chrt)
}
