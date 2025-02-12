package packet

import (
	"errors"

	"github.com/gofrs/uuid"
	"github.com/siyul-park/uniflow/pkg/types"
)

// Packet represents a structured data block exchanged between ports.
type Packet struct {
	id      uuid.UUID
	payload types.Value
}

// None is a predefined packet with no payload.
var None = New(nil)

// ErrDroppedPacket is an error indicating a dropped packet.
var ErrDroppedPacket = types.NewError(errors.New("dropped packet"))

// Join combines multiple packets into one, handling errors and payloads.
func Join(pcks ...*Packet) *Packet {
	if len(pcks) == 0 {
		return None
	} else if len(pcks) == 1 {
		return pcks[0]
	}

	var errs []error
	var payloads []types.Value
	for _, pck := range pcks {
		if pck == nil || pck == None {
			continue
		}

		switch payload := pck.Payload().(type) {
		case types.Error:
			errs = append(errs, payload.Unwrap())
		default:
			payloads = append(payloads, payload)
		}
	}

	if len(errs) > 0 {
		return New(types.NewError(errors.Join(errs...)))
	} else if len(payloads) == 0 {
		return None
	} else if len(payloads) == 1 {
		return New(payloads[0])
	}
	return New(types.NewSlice(payloads...))
}

// New creates a new Packet with the given payload.
func New(payload types.Value) *Packet {
	return &Packet{
		id:      uuid.Must(uuid.NewV7()),
		payload: payload,
	}
}

// ID returns the unique identifier of the packet.
func (p *Packet) ID() uuid.UUID {
	return p.id
}

// Payload returns the data payload of the packet.
func (p *Packet) Payload() types.Value {
	return p.payload
}
