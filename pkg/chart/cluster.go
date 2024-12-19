package chart

import (
	"github.com/gofrs/uuid"
	"sync"

	"github.com/siyul-park/uniflow/pkg/node"
	"github.com/siyul-park/uniflow/pkg/port"
	"github.com/siyul-park/uniflow/pkg/process"
	"github.com/siyul-park/uniflow/pkg/symbol"
)

// ClusterNode manages the ports and symbol table for the cluster.
type ClusterNode struct {
	symbols   []*symbol.Symbol
	table     *symbol.Table
	inPorts   map[string]*port.InPort
	outPorts  map[string]*port.OutPort
	_inPorts  map[string]*port.InPort
	_outPorts map[string]*port.OutPort
	mu        sync.RWMutex
}

var _ node.Node = (*ClusterNode)(nil)

// NewClusterNode creates a new ClusterNode with the provided symbol table.
func NewClusterNode(symbols []*symbol.Symbol, opts ...symbol.TableOption) *ClusterNode {
	return &ClusterNode{
		symbols:   symbols,
		table:     symbol.NewTable(opts...),
		inPorts:   make(map[string]*port.InPort),
		outPorts:  make(map[string]*port.OutPort),
		_inPorts:  make(map[string]*port.InPort),
		_outPorts: make(map[string]*port.OutPort),
	}
}

// Keys returns all keys from the symbol table.
func (n *ClusterNode) Keys() []uuid.UUID {
	keys := make([]uuid.UUID, 0, len(n.symbols))
	for _, sb := range n.symbols {
		keys = append(keys, sb.ID())
	}
	return keys
}

// Lookup retrieves a symbol from the table by its UUID.
func (n *ClusterNode) Lookup(id uuid.UUID) *symbol.Symbol {
	for _, sb := range n.symbols {
		if sb.ID() == id {
			return sb
		}
	}
	return nil
}

// Inbound links an external input to an internal symbol's input port.
func (n *ClusterNode) Inbound(source string, id uuid.UUID, target string) bool {
	n.mu.Lock()
	defer n.mu.Unlock()

	sb := n.Lookup(id)
	if sb == nil {
		return false
	}

	prt := sb.In(target)
	if prt == nil {
		return false
	}

	inPort, ok1 := n.inPorts[source]
	if !ok1 {
		inPort = port.NewIn()
		n.inPorts[source] = inPort
	}

	outPort, ok2 := n._outPorts[source]
	if !ok2 {
		outPort = port.NewOut()
		n._outPorts[source] = outPort
	}

	if !ok1 {
		inPort.AddListener(n.inbound(inPort, outPort))
	}
	if !ok2 {
		outPort.AddListener(n.outbound(inPort, outPort))
	}

	outPort.Link(prt)
	return true
}

// Outbound links an external output to an internal symbol's output port.
func (n *ClusterNode) Outbound(source string, id uuid.UUID, target string) bool {
	n.mu.Lock()
	defer n.mu.Unlock()

	sb := n.Lookup(id)
	if sb == nil {
		return false
	}

	prt := sb.Out(target)
	if prt == nil {
		return false
	}

	inPort, ok1 := n._inPorts[source]
	if !ok1 {
		inPort = port.NewIn()
		n._inPorts[source] = inPort
	}

	outPort, ok2 := n.outPorts[source]
	if !ok2 {
		outPort = port.NewOut()
		n.outPorts[source] = outPort
	}

	if !ok1 {
		inPort.AddListener(n.inbound(inPort, outPort))
	}
	if !ok2 {
		outPort.AddListener(n.outbound(inPort, outPort))
	}

	prt.Link(inPort)
	return true
}

// Load processes all initialization hooks for symbols.
func (n *ClusterNode) Load(hook symbol.LoadHook) error {
	n.mu.Lock()
	defer n.mu.Unlock()

	n.table.AddLoadHook(hook)
	defer n.table.RemoveLoadHook(hook)

	for _, sb := range n.symbols {
		if n.table.Lookup(sb.ID()) != nil {
			continue
		}

		sb := &symbol.Symbol{
			Spec: sb.Spec,
			Node: node.NoCloser(sb.Node),
		}
		if err := n.table.Insert(sb); err != nil {
			return err
		}
	}
	return nil
}

// Unload processes all termination hooks for symbols.
func (n *ClusterNode) Unload(hook symbol.UnloadHook) error {
	n.mu.Lock()
	defer n.mu.Unlock()

	n.table.AddUnloadHook(hook)
	defer n.table.RemoveUnloadHook(hook)

	return n.table.Close()
}

// In returns the input port by name.
func (n *ClusterNode) In(name string) *port.InPort {
	n.mu.RLock()
	defer n.mu.RUnlock()

	return n.inPorts[name]
}

// Out returns the output port by name.
func (n *ClusterNode) Out(name string) *port.OutPort {
	n.mu.RLock()
	defer n.mu.RUnlock()

	return n.outPorts[name]
}

// Close shuts down all ports and the symbol table.
func (n *ClusterNode) Close() error {
	n.mu.RLock()
	defer n.mu.RUnlock()

	if err := n.table.Close(); err != nil {
		return err
	}

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

func (n *ClusterNode) inbound(inPort *port.InPort, outPort *port.OutPort) port.Listener {
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

func (n *ClusterNode) outbound(inPort *port.InPort, outPort *port.OutPort) port.Listener {
	return port.ListenFunc(func(proc *process.Process) {
		reader := inPort.Open(proc)
		writer := outPort.Open(proc)

		for backPck := range writer.Receive() {
			reader.Receive(backPck)
		}
	})
}
