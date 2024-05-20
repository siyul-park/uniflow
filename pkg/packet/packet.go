package packet

import (
	"github.com/siyul-park/uniflow/pkg/primitive"
)

// Packet represents a formalized block of data.
type Packet struct {
	payload primitive.Value
}

var EOF = New(nil)

func Merge(pcks []*Packet) *Packet {
	payloads := make([]primitive.Value, 0, len(pcks))
	for _, pck := range pcks {
		if pck != EOF {
			payloads = append(payloads, pck.Payload())
		}
	}

	if len(payloads) == 0 {
		return EOF
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
