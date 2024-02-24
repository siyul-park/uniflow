package packet

import (
	"github.com/siyul-park/uniflow/pkg/primitive"
)

// Packet represents a formalized block of data.
type Packet struct {
	payload primitive.Value
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
