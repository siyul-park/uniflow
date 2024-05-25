package packet

import (
	"errors"

	"github.com/siyul-park/uniflow/pkg/primitive"
)

// Packet represents a formalized block of data.
type Packet struct {
	payload primitive.Value
}

var None = New(nil)

func Merge(pcks []*Packet) *Packet {
	var errs []error
	for _, pck := range pcks {
		if err, ok := AsError(pck); ok {
			errs = append(errs, err)
		}
	}
	if len(errs) > 0 {
		return NewError(errors.Join(errs...))
	}

	payloads := make([]primitive.Value, 0, len(pcks))
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
		return New(primitive.NewSlice(payloads...))
	}
}

// New creates a new Packet with the given payload.
// It generates a new unique ID for the Packet.
func New(payload primitive.Value) *Packet {
	return &Packet{
		payload: payload,
	}
}

// Payload returns the data payload.
func (p *Packet) Payload() primitive.Value {
	return p.payload
}
