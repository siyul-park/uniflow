package packet

import (
	"github.com/gofrs/uuid"
	"github.com/siyul-park/uniflow/pkg/primitive"
)

// Packet represents a formalized block of data.
type Packet struct {
	id      uuid.UUID
	payload primitive.Value
}

// New creates a new Packet with the given payload.
// It generates a new unique ID for the Packet.
func New(payload primitive.Value) *Packet {
	return &Packet{
		id:      uuid.Must(uuid.NewV7()),
		payload: payload,
	}
}

// ID returns the unique ID.
func (p *Packet) ID() uuid.UUID {
	return p.id
}

// Payload returns the data payload.
func (p *Packet) Payload() primitive.Value {
	return p.payload
}
