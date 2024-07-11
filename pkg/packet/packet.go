package packet

import (
	"errors"

	"github.com/siyul-park/uniflow/pkg/types"
)

// Packet represents a structured data block exchanged between ports.
type Packet struct {
	payload types.Value
}

// None is a predefined packet with no payload.
var None = New(nil)

// ErrDroppedPacket is an error indicating a dropped packet.
var ErrDroppedPacket = errors.New("dropped packet")

// Merge combines multiple packets into one, handling errors and payloads.
func Merge(pcks []*Packet) *Packet {
	var errs []error
	var payloads []types.Value

	for _, pck := range pcks {
		if pck == nil || pck == None {
			continue
		}

		switch payload := pck.Payload().(type) {
		case types.Error:
			errs = append(errs, payload.Interface().(error))
		default:
			payloads = append(payloads, payload)
		}
	}

	if len(errs)+len(payloads) == 0 {
		return None
	}
	if len(errs) > 0 {
		return New(types.NewError(errors.Join(errs...)))
	}
	if len(payloads) == 1 {
		return New(payloads[0])
	}
	return New(types.NewSlice(payloads...))
}

// New creates a new Packet with the given payload.
func New(payload types.Value) *Packet {
	return &Packet{payload: payload}
}

// Payload returns the data payload of the packet.
func (p *Packet) Payload() types.Value {
	return p.payload
}
