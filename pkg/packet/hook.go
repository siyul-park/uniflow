package packet

// Hook defines an interface for handling packets.
type Hook interface {
	Handle(*Packet)
}

// HookFunc is a function type that implements the Handler interface.
type HookFunc func(*Packet)

var _ Hook = HookFunc(nil)

// Handle calls the underlying function represented by HandlerFunc with the provided packet.
func (f HookFunc) Handle(pck *Packet) {
	f(pck)
}
