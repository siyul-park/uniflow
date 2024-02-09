package node

import (
	"github.com/siyul-park/uniflow/pkg/port"
)

// Node represents an operational unit that processes packets.
type Node interface {
	Port(name string) *port.Port // Port retrieves the port with the specified name.
	Close() error                // Close closes the node and releases any resources it holds.
}
