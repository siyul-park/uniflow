package control

import (
	"sync"

	"github.com/siyul-park/uniflow/pkg/node"
	"github.com/siyul-park/uniflow/pkg/object"
	"github.com/siyul-park/uniflow/pkg/packet"
	"github.com/siyul-park/uniflow/pkg/port"
	"github.com/siyul-park/uniflow/pkg/process"
	"github.com/siyul-park/uniflow/pkg/scheme"
	"github.com/siyul-park/uniflow/pkg/spec"
)

// ForkNode is a node that forks processes and manages packet forwarding between ports.
type ForkNode struct {
	inPort  *port.InPort
	outPort *port.OutPort
	errPort *port.OutPort
	mu      sync.RWMutex
}

// ForkNodeSpec holds the specifications for creating a ForkNode.
type ForkNodeSpec struct {
	spec.Meta `map:",inline"`
}

const KindFork = "fork"

var _ node.Node = (*ForkNode)(nil)

// NewForkNode creates a new ForkNode.
func NewForkNode() *ForkNode {
	n := &ForkNode{
		inPort:  port.NewIn(),
		outPort: port.NewOut(),
		errPort: port.NewOut(),
	}

	n.inPort.Accept(port.ListenFunc(n.forward))
	n.outPort.Accept(port.ListenFunc(n.backward))
	n.errPort.Accept(port.ListenFunc(n.catch))

	return n
}

// In returns the input port with the specified name.
func (n *ForkNode) In(name string) *port.InPort {
	n.mu.RLock()
	defer n.mu.RUnlock()

	switch name {
	case node.PortIn:
		return n.inPort
	default:
	}

	return nil
}

// Out returns the output port with the specified name.
func (n *ForkNode) Out(name string) *port.OutPort {
	n.mu.RLock()
	defer n.mu.RUnlock()

	switch name {
	case node.PortOut:
		return n.outPort
	case node.PortErr:
		return n.errPort
	default:
	}

	return nil
}

// Close closes all ports associated with the node.
func (n *ForkNode) Close() error {
	n.mu.Lock()
	defer n.mu.Unlock()

	n.inPort.Close()
	n.outPort.Close()
	n.errPort.Close()

	return nil
}

func (n *ForkNode) forward(proc *process.Process) {
	n.mu.RLock()
	defer n.mu.RUnlock()

	inReader := n.inPort.Open(proc)

	for {
		inPck, ok := <-inReader.Read()
		if !ok {
			return
		}

		child := proc.Fork()
		outWriter := n.outPort.Open(child)

		outWriter.Write(inPck)
		inReader.Receive(packet.None)
	}
}

func (n *ForkNode) backward(proc *process.Process) {
	n.mu.RLock()
	defer n.mu.RUnlock()

	outWriter := n.outPort.Open(proc)
	errWriter := n.errPort.Open(proc)

	for {
		backPck, ok := <-outWriter.Receive()
		if !ok {
			return
		}

		if err, ok := backPck.Payload().(object.Error); ok {
			if errWriter.Write(backPck) == 0 {
				proc.Wait()
				proc.Exit(err)
			}
		} else {
			proc.Wait()
			proc.Exit(nil)
		}
	}
}

func (n *ForkNode) catch(proc *process.Process) {
	n.mu.RLock()
	defer n.mu.RUnlock()

	errWriter := n.errPort.Open(proc)

	for {
		backPck, ok := <-errWriter.Receive()
		if !ok {
			return
		}

		err, _ := backPck.Payload().(object.Error)

		proc.Wait()
		proc.Exit(err)
	}
}

// NewForkNodeCodec creates a new codec for ForkNodeSpec.
func NewForkNodeCodec() scheme.Codec {
	return scheme.CodecWithType(func(_ *ForkNodeSpec) (node.Node, error) {
		return NewForkNode(), nil
	})
}
