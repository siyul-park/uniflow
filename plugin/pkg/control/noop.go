package control

import (
	"sync"

	"github.com/siyul-park/uniflow/pkg/node"
	"github.com/siyul-park/uniflow/pkg/port"
	"github.com/siyul-park/uniflow/pkg/process"
	"github.com/siyul-park/uniflow/pkg/scheme"
)

// NoOpNode represents a node that performs no operation and simply forwards incoming packets.
type NoOpNode struct {
	inPort *port.InPort
	mu     sync.RWMutex
}

// NoOpNodeSpec defines the specification for creating a NoOpNode.
type NoOpNodeSpec struct {
	scheme.SpecMeta `map:",inline"`
}

var _ node.Node = (*NoOpNode)(nil)

const KindNoOp = "noop"

// NewNoOpNode creates a new instance of NoOpNode.
func NewNoOpNode() *NoOpNode {
	n := &NoOpNode{
		inPort: port.NewIn(),
	}

	n.inPort.AddHandler(port.HandlerFunc(n.forward))

	return n
}

// In returns the input port with the specified name.
func (n *NoOpNode) In(name string) *port.InPort {
	n.mu.RLock()
	defer n.mu.RUnlock()

	switch name {
	case node.PortIn:
		return n.inPort
	default:
	}

	return nil
}

// Out returns nil as NoOpNode does not have any output port.
func (n *NoOpNode) Out(name string) *port.OutPort {
	return nil
}

// Close closes all ports associated with the node.
func (n *NoOpNode) Close() error {
	n.mu.Lock()
	defer n.mu.Unlock()

	n.inPort.Close()

	return nil
}

// forward forwards incoming packets.
func (n *NoOpNode) forward(proc *process.Process) {
	n.mu.RLock()
	defer n.mu.RUnlock()

	inReader := n.inPort.Open(proc)

	for {
		inPck, ok := <-inReader.Read()
		if !ok {
			return
		}

		proc.Stack().Clear(inPck)
	}
}

// NewNoOpNodeCodec creates a codec for decoding NoOpNodeSpec.
func NewNoOpNodeCodec() scheme.Codec {
	return scheme.CodecWithType(func(spec *NoOpNodeSpec) (node.Node, error) {
		return NewNoOpNode(), nil
	})
}
