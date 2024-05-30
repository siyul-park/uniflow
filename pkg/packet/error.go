package packet

import (
	"errors"

	"github.com/siyul-park/uniflow/pkg/object"
)

var (
	ErrInvalidPacket = errors.New("invalid packet")
	ErrDiscardPacket = errors.New("discarded packet")
)

// WithError creates a new Packet representing an error with the given error.
// FIXME: support object.Error
func WithError(err error) *Packet {
	pairs := []object.Object{
		object.NewString("__error"),
		object.True,
		object.NewString("error"),
		object.NewString(err.Error()),
	}
	return New(object.NewMap(pairs...))
}

// AsError extracts the error from a Packet, returning it along with a boolean indicating whether
// the Packet contains error information. If the Packet does not represent an error, the
// returned error is nil, and the boolean is false.
// FIXME: support object.Error
func AsError(pck *Packet) (error, bool) {
	if pck == nil {
		return nil, false
	}

	payload := pck.Payload()

	if ok, _ := object.Pick[bool](payload, "__error"); !ok {
		return nil, false
	}

	if err, ok := object.Pick[string](payload, "error"); !ok {
		return nil, false
	} else {
		return errors.New(err), true
	}
}
