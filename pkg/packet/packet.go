package packet

import (
	"github.com/oklog/ulid/v2"
	"github.com/siyul-park/uniflow/pkg/primitive"
)

// Packet represents a formalized block of data.
type Packet struct {
	id      ulid.ULID
	payload primitive.Value
}

// New creates a new Packet with the given payload.
// It generates a new unique ID for the Packet.
func New(payload primitive.Value) *Packet {
	return &Packet{
		id:      ulid.Make(),
		payload: payload,
	}
}

// ID returns the unique identifier (ID) of the Packet.
func (p *Packet) ID() ulid.ULID {
	return p.id
}

// Payload returns the data payload of the Packet.
func (p *Packet) Payload() primitive.Value {
	return p.payload
}
