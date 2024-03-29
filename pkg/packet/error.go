package packet

import (
	"errors"

	"github.com/siyul-park/uniflow/pkg/primitive"
)

var (
	ErrInvalidPacket = errors.New("invalid packet")
	ErrDiscardPacket = errors.New("discarded packet")
)

// WithError creates a new Packet representing an error with the given error and optional cause.
// It constructs a Packet with error details, including the error message.
// If a cause is provided, it is attached to the error packet.
func WithError(err error, cause *Packet) *Packet {
	pairs := []primitive.Value{
		primitive.NewString("__error"),
		primitive.TRUE,
		primitive.NewString("error"),
		primitive.NewString(err.Error()),
	}

	if cause != nil {
		pairs = append(pairs, primitive.NewString("cause"), cause.Payload())
	}

	return New(primitive.NewMap(pairs...))
}

// AsError extracts the error from a Packet, returning it along with a boolean indicating whether
// the Packet contains error information. If the Packet does not represent an error, the
// returned error is nil, and the boolean is false.
func AsError(pck *Packet) (error, bool) {
	payload := pck.Payload()

	if ok, _ := primitive.Pick[bool](payload, "__error"); !ok {
		return nil, false
	}
	if err, ok := primitive.Pick[string](payload, "error"); !ok {
		return nil, false
	} else {
		return errors.New(err), true
	}
}
