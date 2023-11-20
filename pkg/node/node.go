package node

import (
	"github.com/oklog/ulid/v2"
	"github.com/siyul-park/uniflow/pkg/port"
)

type (
	// Node is an operational unit that processes *packet.Packet.
	Node interface {
		ID() ulid.ULID
		Port(name string) (*port.Port, bool)
		Close() error
	}
)
