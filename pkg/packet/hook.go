package packet

// Hook defines an interface for handling packets.
type Hook interface {
	// Handle processes the given packet.
	Handle(*Packet)
}

type hook struct {
	handle func(*Packet)
}

var _ Hook = (*hook)(nil)

// HookFunc creates a new Hook from the provided function.
func HookFunc(handle func(*Packet)) Hook {
	return &hook{handle: handle}
}

func (h *hook) Handle(pck *Packet) {
	h.handle(pck)
}
