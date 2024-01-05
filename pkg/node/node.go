package node

import (
	"github.com/siyul-park/uniflow/pkg/port"
)

// Node is an operational unit that processes *packet.Packet.
type Node interface {
	Port(name string) (*port.Port, bool)
	Close() error
}
