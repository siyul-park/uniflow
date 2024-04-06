package control

import (
	"sync"

	"github.com/siyul-park/uniflow/pkg/node"
	"github.com/siyul-park/uniflow/pkg/packet"
	"github.com/siyul-park/uniflow/pkg/port"
	"github.com/siyul-park/uniflow/pkg/process"
	"github.com/siyul-park/uniflow/pkg/scheme"
)

// JumpNode redirects packets from the input port to the intermediate port for processing by connected nodes, then outputs the results to the output port.
type JumpNode struct {
	inPort  *port.InPort
	outPort *port.OutPort
	ioPort  *port.OutPort
	errPort *port.OutPort
	mu      sync.RWMutex
}

// JumpNodeSpec holds the specifications for creating a JumpNode.
type JumpNodeSpec struct {
	scheme.SpecMeta `map:",inline"`
}

var _ node.Node = (*JumpNode)(nil)

const KindJump = "jump"

// NewJumpNode creates a new JumpNode.
func NewJumpNode() *JumpNode {
	n := &JumpNode{
		inPort:  port.NewIn(),
		outPort: port.NewOut(),
		ioPort:  port.NewOut(),
		errPort: port.NewOut(),
	}

	n.ioPort.AddHandler(port.HandlerFunc(n.turn))
	n.inPort.AddHandler(port.HandlerFunc(n.forward))
	n.outPort.AddHandler(port.HandlerFunc(n.backward))
	n.errPort.AddHandler(port.HandlerFunc(n.catch))

	return n
}

// In returns the input port with the specified name.
func (n *JumpNode) In(name string) *port.InPort {
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
func (n *JumpNode) Out(name string) *port.OutPort {
	n.mu.RLock()
	defer n.mu.RUnlock()

	switch name {
	case node.PortOut:
		return n.outPort
	case node.PortIO:
		return n.ioPort
	case node.PortErr:
		return n.errPort
	default:
	}

	return nil
}

// Close closes all ports associated with the node.
func (n *JumpNode) Close() error {
	n.mu.Lock()
	defer n.mu.Unlock()

	n.inPort.Close()
	n.outPort.Close()
	n.ioPort.Close()
	n.errPort.Close()

	return nil
}

func (n *JumpNode) turn(proc *process.Process) {
	n.mu.RLock()
	defer n.mu.RUnlock()

	ioWriter := n.ioPort.Open(proc)
	outWriter := n.outPort.Open(proc)

	for {
		backPck, ok := <-ioWriter.Receive()
		if !ok {
			return
		}

		if _, ok := packet.AsError(backPck); ok {
			n.throw(proc, backPck)
		} else {
			outWriter.Write(backPck)
		}
	}
}

func (n *JumpNode) forward(proc *process.Process) {
	n.mu.RLock()
	defer n.mu.RUnlock()

	inReader := n.inPort.Open(proc)
	ioWriter := n.ioPort.Open(proc)

	for {
		inPck, ok := <-inReader.Read()
		if !ok {
			return
		}

		ioWriter.Write(inPck)
	}
}

func (n *JumpNode) backward(proc *process.Process) {
	n.mu.RLock()
	defer n.mu.RUnlock()

	inReader := n.inPort.Open(proc)
	outWriter := n.outPort.Open(proc)

	for {
		backPck, ok := <-outWriter.Receive()
		if !ok {
			return
		}

		inReader.Receive(backPck)
	}
}

func (n *JumpNode) catch(proc *process.Process) {
	n.mu.RLock()
	defer n.mu.RUnlock()

	inReader := n.inPort.Open(proc)
	errWriter := n.errPort.Open(proc)

	for {
		backPck, ok := <-errWriter.Receive()
		if !ok {
			return
		}

		inReader.Receive(backPck)
	}
}

func (n *JumpNode) throw(proc *process.Process, errPck *packet.Packet) {
	inReader := n.inPort.Open(proc)
	errWriter := n.errPort.Open(proc)

	if errWriter.Links() > 0 {
		errWriter.Write(errPck)
	} else {
		inReader.Receive(errPck)
	}
}

// NewJumpNodeCodec creates a new codec for JumpNodeSpec.
func NewJumpNodeCodec() scheme.Codec {
	return scheme.CodecWithType[*JumpNodeSpec](func(spec *JumpNodeSpec) (node.Node, error) {
		return NewJumpNode(), nil
	})
}
