package packet

// Handler defines an interface for handling packets.
type Handler interface {
	Handle(*Packet)
}

// HandlerFunc is a function type that implements the Handler interface.
type HandlerFunc func(*Packet)

var _ Handler = HandlerFunc(nil)

// Handle calls the HandlerFunc with the provided packet.
func (f HandlerFunc) Handle(pck *Packet) {
	f(pck)
}
