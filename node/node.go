package node

import (
	"github.com/siyul-park/uniflow/port"
)

// Node represents an operational unit that processes packets.
type Node interface {
	In(name string) *port.InPort
	Out(name string) *port.OutPort
	Close() error
}
