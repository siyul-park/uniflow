package chart

// UnlinkHook defines an interface for handling the unloading of a chart.
type UnlinkHook interface {
	// Unlink is called when a chart is unloaded and may return an error.
	Unlink(*Chart) error
}

// UnlinkHooks is a slice of UnloadHook, processed in reverse order.
type UnlinkHooks []UnlinkHook

// unlinkHook wraps an unload function to implement UnloadHook.
type unlinkHook struct {
	unlink func(*Chart) error
}

var _ UnlinkHook = (UnlinkHooks)(nil)
var _ UnlinkHook = (*unlinkHook)(nil)

// UnlinkFunc creates an UnloadHook from the given function.
func UnlinkFunc(unlink func(*Chart) error) UnlinkHook {
	return &unlinkHook{unlink: unlink}
}

func (h UnlinkHooks) Unlink(chrt *Chart) error {
	for i := len(h) - 1; i >= 0; i-- {
		hook := h[i]
		if err := hook.Unlink(chrt); err != nil {
			return err
		}
	}
	return nil
}

func (h *unlinkHook) Unlink(chrt *Chart) error {
	return h.unlink(chrt)
}
