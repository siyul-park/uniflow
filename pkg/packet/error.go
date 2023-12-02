package packet

import "github.com/siyul-park/uniflow/pkg/primitive"

// NewError creates a new Packet to represent an error.
// It takes an error and an optional cause Packet and constructs a Packet with error details.
func NewError(err error, cause *Packet) *Packet {
	pairs := []primitive.Value{
		primitive.NewString("error"),
		primitive.NewString(err.Error()),
	}

	if cause != nil {
		pairs = append(pairs, primitive.NewString("cause"), cause.Payload())
	}

	return New(primitive.NewMap(pairs...))
}

// IsError checks if the given Packet represents an error.
func IsError(pck *Packet) bool {
	payload := pck.Payload()
	_, ok := primitive.Pick[string](payload, "error")
	return ok
}
