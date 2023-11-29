package packet

import (
	"github.com/oklog/ulid/v2"
	"github.com/siyul-park/uniflow/pkg/primitive"
)

type (
	// Packet is a formalized block of data.
	Packet struct {
		id      ulid.ULID
		payload primitive.Value
	}
)

// NewError return a new Packet for error.
func NewError(err error, cause *Packet) *Packet {
	var pairs []primitive.Value
	pairs = append(pairs, primitive.NewString("error"), primitive.NewString(err.Error()))
	if cause != nil {
		pairs = append(pairs, primitive.NewString("cause"), cause.Payload())
	}

	return New(primitive.NewMap(pairs...))
}

// New returns a new Packet.
func New(payload primitive.Value) *Packet {
	return &Packet{
		id:      ulid.Make(),
		payload: payload,
	}
}

// ID returns the ID of the Packet
func (pck *Packet) ID() ulid.ULID {
	return pck.id
}

// Payload returns the payload of the Packet.
func (pck *Packet) Payload() primitive.Value {
	return pck.payload
}
