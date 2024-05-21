package control

import (
	"sync"

	"github.com/siyul-park/uniflow/pkg/node"
	"github.com/siyul-park/uniflow/pkg/port"
	"github.com/siyul-park/uniflow/pkg/process"
	"github.com/siyul-park/uniflow/pkg/scheme"
)

// NOPNode represents a node that performs no operation and simply forwards incoming packets.
type NOPNode struct {
	inPort *port.InPort
	mu     sync.RWMutex
}

// NOPNodeSpec defines the specification for creating a NOPNode.
type NOPNodeSpec struct {
	scheme.SpecMeta `map:",inline"`
}

var _ node.Node = (*NOPNode)(nil)

const KindNOP = "nop"

// NewNOPNode creates a new instance of NOPNode.
func NewNOPNode() *NOPNode {
	n := &NOPNode{
		inPort: port.NewIn(),
	}

	n.inPort.AddInitHook(port.InitHookFunc(n.forward))

	return n
}

// In returns the input port with the specified name.
func (n *NOPNode) In(name string) *port.InPort {
	n.mu.RLock()
	defer n.mu.RUnlock()

	switch name {
	case node.PortIn:
		return n.inPort
	default:
	}

	return nil
}

// Out returns nil as NOPNode does not have any output port.
func (n *NOPNode) Out(name string) *port.OutPort {
	return nil
}

// Close closes all ports associated with the node.
func (n *NOPNode) Close() error {
	n.mu.Lock()
	defer n.mu.Unlock()

	n.inPort.Close()

	return nil
}

// forward forwards incoming packets.
func (n *NOPNode) forward(proc *process.Process) {
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

// NewNOPNodeCodec creates a codec for decoding NOPNodeSpec.
func NewNOPNodeCodec() scheme.Codec {
	return scheme.CodecWithType(func(spec *NOPNodeSpec) (node.Node, error) {
		return NewNOPNode(), nil
	})
}
