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
	table     *symbol.Table
	inPorts   map[string]*port.InPort
	outPorts  map[string]*port.OutPort
	_inPorts  map[string]*port.InPort
	_outPorts map[string]*port.OutPort
	mu        sync.RWMutex
}

var _ node.Node = (*ClusterNode)(nil)

// NewClusterNode creates a new ClusterNode with the provided symbol table.
func NewClusterNode(table *symbol.Table) *ClusterNode {
	return &ClusterNode{
		table:     table,
		inPorts:   make(map[string]*port.InPort),
		outPorts:  make(map[string]*port.OutPort),
		_inPorts:  make(map[string]*port.InPort),
		_outPorts: make(map[string]*port.OutPort),
	}
}

// Inbound sets up an input port and links it to the provided port.
func (n *ClusterNode) Inbound(source string, id uuid.UUID, target string) bool {
	n.mu.Lock()
	defer n.mu.Unlock()

	sb := n.table.Lookup(id)
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

// Outbound sets up an output port and links it to the provided port.
func (n *ClusterNode) Outbound(source string, id uuid.UUID, target string) bool {
	n.mu.Lock()
	defer n.mu.Unlock()

	sb := n.table.Lookup(id)
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

// Symbols retrieves all symbols from the table.
func (n *ClusterNode) Symbols() []*symbol.Symbol {
	n.mu.RLock()
	defer n.mu.RUnlock()

	var symbols []*symbol.Symbol
	for _, key := range n.table.Keys() {
		if sym := n.table.Lookup(key); sym != nil {
			symbols = append(symbols, sym)
		}
	}
	return symbols
}

// Symbol retrieves a specific symbol by UUID.
func (n *ClusterNode) Symbol(id uuid.UUID) *symbol.Symbol {
	n.mu.RLock()
	defer n.mu.RUnlock()

	return n.table.Lookup(id)
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
