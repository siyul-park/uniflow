package control

import (
	"github.com/siyul-park/uniflow/pkg/node"
	"github.com/siyul-park/uniflow/pkg/port"
	"github.com/siyul-park/uniflow/pkg/process"
	"github.com/siyul-park/uniflow/pkg/scheme"
	"github.com/siyul-park/uniflow/pkg/spec"
	"github.com/siyul-park/uniflow/pkg/symbol"
	"sync"
)

// BlockNodeSpec defines the specification for creating a BlockNode.
type BlockNodeSpec struct {
	spec.Meta `map:",inline"`
	Specs     []*spec.Unstructured `map:"specs"`
}

// BlockNode systematically manages complex data processing flows and executes multiple sub-nodes sequentially.
type BlockNode struct {
	symbols   []*symbol.Symbol
	inPorts   map[string]*port.InPort
	outPorts  map[string]*port.OutPort
	_inPorts  map[string]*port.InPort
	_outPorts map[string]*port.OutPort
	mu        sync.RWMutex
}

const KindBlock = "block"

var _ node.Node = (*BlockNode)(nil)

// NewBlockNodeCodec creates a new codec for BlockNodeSpec.
func NewBlockNodeCodec(s *scheme.Scheme) scheme.Codec {
	return scheme.CodecWithType(func(spec *BlockNodeSpec) (node.Node, error) {
		symbols := make([]*symbol.Symbol, 0, len(spec.Specs))
		for _, sp := range spec.Specs {
			decoded, err := s.Decode(sp)
			if err != nil {
				for _, n := range symbols {
					n.Close()
				}
				return nil, err
			}
			n, err := s.Compile(decoded)
			if err != nil {
				for _, n := range symbols {
					n.Close()
				}
				return nil, err
			}
			symbols = append(symbols, &symbol.Symbol{
				Spec: decoded,
				Node: n,
			})
		}
		return NewBlockNode(symbols...), nil
	})
}

// NewBlockNode creates a new BlockNode with the specified symbols.
func NewBlockNode(nodes ...*symbol.Symbol) *BlockNode {
	n := &BlockNode{
		symbols:   nodes,
		inPorts:   make(map[string]*port.InPort),
		outPorts:  make(map[string]*port.OutPort),
		_inPorts:  make(map[string]*port.InPort),
		_outPorts: make(map[string]*port.OutPort),
	}

	for i := 1; i < len(n.symbols); i++ {
		outPort := n.symbols[i-1].Out(node.PortOut)
		inPort := n.symbols[i].In(node.PortIn)
		if outPort == nil || inPort == nil {
			break
		}
		outPort.Link(inPort)
	}

	if len(n.symbols) > 1 {
		inPort := port.NewIn()
		outPort := port.NewOut()

		n._inPorts[node.PortErr] = inPort
		n.outPorts[node.PortErr] = outPort

		for _, cur := range n.symbols {
			if err := cur.Out(node.PortErr); err != nil {
				err.Link(inPort)
			}
		}

		inPort.AddListener(n.inbound(inPort, outPort))
		outPort.AddListener(n.outbound(inPort, outPort))
	}

	return n
}

// Load iterates over nodes in reverse order, invoking hook.Load for each node.
func (n *BlockNode) Load(hook symbol.LoadHook) error {
	n.mu.RLock()
	defer n.mu.RUnlock()

	for i := len(n.symbols) - 1; i >= 0; i-- {
		sb := n.symbols[i]
		if err := hook.Load(sb); err != nil {
			return err
		}
	}
	return nil
}

// Unload iterates over nodes, invoking hook.Unload for each node.
func (n *BlockNode) Unload(hook symbol.UnloadHook) error {
	n.mu.RLock()
	defer n.mu.RUnlock()

	for _, sb := range n.symbols {
		if err := hook.Unload(sb); err != nil {
			return err
		}
	}
	return nil
}

// In returns the input port with the specified name.
func (n *BlockNode) In(name string) *port.InPort {
	n.mu.Lock()
	defer n.mu.Unlock()

	if inPort, ok := n.inPorts[name]; ok {
		return inPort
	}
	if len(n.symbols) > 0 {
		if inPort := n.symbols[0].In(name); inPort != nil {
			n.inPorts[name] = inPort
			return inPort
		}
	}
	return nil
}

// Out returns the output port with the specified name.
func (n *BlockNode) Out(name string) *port.OutPort {
	n.mu.Lock()
	defer n.mu.Unlock()

	if outPort, ok := n.outPorts[name]; ok {
		return outPort
	}
	if len(n.symbols) > 0 {
		if outPort := n.symbols[len(n.symbols)-1].Out(name); outPort != nil {
			n.outPorts[name] = outPort
			return outPort
		}
	}
	return nil
}

// Close closes all ports and sub-nodes associated with the BlockNode.
func (n *BlockNode) Close() error {
	n.mu.RLock()
	defer n.mu.RUnlock()

	for _, sb := range n.symbols {
		if err := sb.Close(); err != nil {
			return err
		}
	}
	for _, inPort := range n.inPorts {
		inPort.Close()
	}
	for _, inPort := range n._inPorts {
		inPort.Close()
	}
	for _, outPort := range n.outPorts {
		outPort.Close()
	}
	for _, outPort := range n._outPorts {
		outPort.Close()
	}
	return nil
}

func (n *BlockNode) inbound(inPort *port.InPort, outPort *port.OutPort) port.Listener {
	return port.ListenFunc(func(proc *process.Process) {
		reader := inPort.Open(proc)
		writer := outPort.Open(proc)

		for inPck := range reader.Read() {
			if writer.Write(inPck) == 0 {
				reader.Receive(inPck)
			}
		}
	})
}

func (n *BlockNode) outbound(inPort *port.InPort, outPort *port.OutPort) port.Listener {
	return port.ListenFunc(func(proc *process.Process) {
		reader := inPort.Open(proc)
		writer := outPort.Open(proc)

		for backPck := range writer.Receive() {
			reader.Receive(backPck)
		}
	})
}
