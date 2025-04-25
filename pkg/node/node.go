package node

import "github.com/siyul-park/uniflow/pkg/port"

// Node represents a unit that processes packets with input and output ports.
type Node interface {
	In(name string) *port.InPort   // Returns the input port by name.
	Out(name string) *port.OutPort // Returns the output port by name.
	Close() error                  // Closes the node and returns any error.
}
