package node

import (
	"github.com/oklog/ulid/v2"
	"github.com/siyul-park/uniflow/pkg/port"
)

// Node is an operational unit that processes *packet.Packet.
type Node interface {
	ID() ulid.ULID
	Port(name string) (*port.Port, bool)
	Close() error
}

const (
	PortIO  = "io"
	PortIn  = "in"
	PortOut = "out"
	PortErr = "error"
)
