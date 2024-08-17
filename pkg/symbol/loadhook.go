package symbol

// LoadHook defines an interface for handling events when a symbol is loaded.
type LoadHook interface {
	Load(*Symbol) error
}

type loadHook struct {
	load func(*Symbol) error
}

var _ LoadHook = (*loadHook)(nil)

// LoadFunc creates a new LoadHook from the provided function.
func LoadFunc(load func(*Symbol) error) LoadHook {
	return &loadHook{load: load}
}

func (h *loadHook) Load(sym *Symbol) error {
	return h.load(sym)
}
