package packet

import (
	"errors"

	"github.com/siyul-park/uniflow/pkg/types"
)

// Packet represents a structured data block exchanged between ports.
type Packet struct {
	payload types.Value
}

// None is a special empty packet.
var None = New(nil)

// Merge combines multiple packets into one, handling errors and payloads.
func Merge(pcks []*Packet) *Packet {
	var errs []error
	var payloads []types.Value

	for _, pck := range pcks {
		if pck == nil || pck == None {
			continue
		}

		if err, ok := pck.Payload().(types.Error); ok {
			errs = append(errs, err.Interface().(error))
		} else {
			payloads = append(payloads, pck.Payload())
		}
	}

	if len(errs)+len(payloads) == 0 {
		return None
	}

	if len(errs) == 1 {
		return New(types.NewError(errs[0]))
	} else if len(errs) > 1 {
		return New(types.NewError(errors.Join(errs...)))
	}

	if len(payloads) == 1 {
		return New(payloads[0])
	} else {
		return New(types.NewSlice(payloads...))
	}
}

// New creates a new Packet with the given payload.
func New(payload types.Value) *Packet {
	return &Packet{
		payload: payload,
	}
}

// Payload returns the data payload of the packet.
func (p *Packet) Payload() types.Value {
	return p.payload
}
