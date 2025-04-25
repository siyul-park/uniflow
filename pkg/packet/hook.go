package packet

// Hook defines an interface for processing packets.
type Hook interface {
	// Handle processes the specified packet.
	Handle(*Packet)
}

// Hooks is a slice of Hook interfaces, allowing multiple Hooks to be handled together.
type Hooks []Hook

type hook struct {
	handle func(*Packet)
}

var (
	_ Hook = (Hooks)(nil)
	_ Hook = (*hook)(nil)
)

// HookFunc creates a new Hook using the provided function.
func HookFunc(handle func(*Packet)) Hook {
	return &hook{handle: handle}
}

// Handle processes each packet using the Hooks in the slice.
func (h Hooks) Handle(pck *Packet) {
	for _, hook := range h {
		hook.Handle(pck)
	}
}

func (h *hook) Handle(pck *Packet) {
	h.handle(pck)
}
