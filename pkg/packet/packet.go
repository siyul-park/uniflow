package packet

import (
	"errors"

	"github.com/siyul-park/uniflow/pkg/object"
)

// Packet represents a formalized block of data.
type Packet struct {
	payload object.Object
}

var None = New(nil)

func Merge(pcks []*Packet) *Packet {
	if len(pcks) == 0 {
		return None
	} else if len(pcks) == 1 {
		return pcks[0]
	}

	var errs []error
	for _, pck := range pcks {
		if err, ok := pck.Payload().(*object.Error); ok {
			errs = append(errs, err.Interface().(error))
		}
	}
	if len(errs) == 1 {
		return New(object.NewError(errs[0]))
	} else if len(errs) > 1 {
		return New(object.NewError(errors.Join(errs...)))
	}

	payloads := make([]object.Object, 0, len(pcks))
	for _, pck := range pcks {
		if pck != nil && pck != None {
			payloads = append(payloads, pck.Payload())
		}
	}

	if len(payloads) == 0 {
		return None
	} else if len(payloads) == 1 {
		return New(payloads[0])
	} else {
		return New(object.NewSlice(payloads...))
	}
}

// New creates a new Packet with the given payload.
// It generates a new unique ID for the Packet.
func New(payload object.Object) *Packet {
	return &Packet{
		payload: payload,
	}
}

// Payload returns the data payload.
func (p *Packet) Payload() object.Object {
	return p.payload
}
