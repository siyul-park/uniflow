package node

import "github.com/siyul-park/uniflow/pkg/port"

// Node represents a unit that processes packets with input and output ports.
type Node interface {
	// In returns the input port identified by 'name'.
	In(name string) *port.InPort
	// Out returns the output port identified by 'name'.
	Out(name string) *port.OutPort
	// Close terminates the node and returns any encountered error.
	Close() error
}
