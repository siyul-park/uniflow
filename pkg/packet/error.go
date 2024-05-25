package packet

import (
	"errors"

	"github.com/siyul-park/uniflow/pkg/primitive"
)

var (
	ErrInvalidPacket = errors.New("invalid packet")
	ErrDiscardPacket = errors.New("discarded packet")
)

// NewError creates a new Packet representing an error with the given error.
func NewError(err error) *Packet {
	pairs := []primitive.Value{
		primitive.NewString("__error"),
		primitive.TRUE,
		primitive.NewString("error"),
		primitive.NewString(err.Error()),
	}
	return New(primitive.NewMap(pairs...))
}

// AsError extracts the error from a Packet, returning it along with a boolean indicating whether
// the Packet contains error information. If the Packet does not represent an error, the
// returned error is nil, and the boolean is false.
func AsError(pck *Packet) (error, bool) {
	if pck == nil {
		return nil, false
	}

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
