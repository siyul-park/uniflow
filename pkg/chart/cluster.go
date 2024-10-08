package chart

import (
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
func (n *ClusterNode) Inbound(name string, prt *port.InPort) {
	n.mu.Lock()
	defer n.mu.Unlock()

	inPort := port.NewIn()
	outPort := port.NewOut()

	n.inPorts[node.PortErr] = inPort
	n._outPorts[node.PortErr] = outPort

	outPort.Link(prt)

	inPort.AddListener(n.inbound(inPort, outPort))
	outPort.AddListener(n.outbound(inPort, outPort))
}

// Outbound sets up an output port and links it to the provided port.
func (n *ClusterNode) Outbound(name string, prt *port.OutPort) {
	n.mu.Lock()
	defer n.mu.Unlock()

	inPort := port.NewIn()
	outPort := port.NewOut()

	n._inPorts[node.PortErr] = inPort
	n.outPorts[node.PortErr] = outPort

	prt.Link(inPort)

	inPort.AddListener(n.inbound(inPort, outPort))
	outPort.AddListener(n.outbound(inPort, outPort))
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
