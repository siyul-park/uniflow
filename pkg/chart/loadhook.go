package chart

// LoadHook defines an interface for handling the loading of a chart.
type LoadHook interface {
	// Load processes the loading of a chart and may return an error.
	Load(*Chart) error
}

type LoadHooks []LoadHook

type loadHook struct {
	load func(*Chart) error
}

var _ LoadHook = (LoadHooks)(nil)
var _ LoadHook = (*loadHook)(nil)

// LoadFunc creates a LoadHook from the given function.
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
