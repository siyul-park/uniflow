package chart

// LinkHook defines an interface for handling the loading of a chart.
type LinkHook interface {
	// Link processes the loading of a chart and may return an error.
	Link(*Chart) error
}

type LinkHooks []LinkHook

type linkHook struct {
	link func(*Chart) error
}

var _ LinkHook = (LinkHooks)(nil)
var _ LinkHook = (*linkHook)(nil)

// LinkFunc creates a LoadHook from the given function.
func LinkFunc(link func(*Chart) error) LinkHook {
	return &linkHook{link: link}
}

func (h LinkHooks) Link(chrt *Chart) error {
	for _, hook := range h {
		if err := hook.Link(chrt); err != nil {
			return err
		}
	}
	return nil
}

func (h *linkHook) Link(chrt *Chart) error {
	return h.link(chrt)
}
