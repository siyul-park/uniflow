package control

import (
	"github.com/siyul-park/uniflow/pkg/node"
	"github.com/siyul-park/uniflow/pkg/packet"
	"github.com/siyul-park/uniflow/pkg/port"
	"github.com/siyul-park/uniflow/pkg/process"
	"github.com/siyul-park/uniflow/pkg/scheme"
	"github.com/siyul-park/uniflow/pkg/spec"
	"github.com/siyul-park/uniflow/pkg/types"
)

// ForkNodeSpec defines the specification for creating a ForkNode.
type ForkNodeSpec struct {
	spec.Meta `json:",inline"`
}

// ForkNode asynchronously branches the data processing flow into separate processes.
type ForkNode struct {
	inPort  *port.InPort
	outPort *port.OutPort
	errPort *port.OutPort
}

const KindFork = "fork"

var _ node.Node = (*ForkNode)(nil)

// NewForkNodeCodec creates a new codec for ForkNodeSpec.
func NewForkNodeCodec() scheme.Codec {
	return scheme.CodecWithType(func(_ *ForkNodeSpec) (node.Node, error) {
		return NewForkNode(), nil
	})
}

// NewForkNode creates a new ForkNode.
func NewForkNode() *ForkNode {
	n := &ForkNode{
		inPort:  port.NewIn(),
		outPort: port.NewOut(),
		errPort: port.NewOut(),
	}

	n.inPort.AddListener(port.ListenFunc(n.forward))
	n.outPort.AddListener(port.ListenFunc(n.backward))
	n.errPort.AddListener(port.ListenFunc(n.catch))

	return n
}

// In returns the input port with the specified name.
func (n *ForkNode) In(name string) *port.InPort {
	switch name {
	case node.PortIn:
		return n.inPort
	default:
		return nil
	}
}

// Out returns the output port with the specified name.
func (n *ForkNode) Out(name string) *port.OutPort {
	switch name {
	case node.PortOut:
		return n.outPort
	case node.PortError:
		return n.errPort
	default:
		return nil
	}
}

// Close closes all ports associated with the node.
func (n *ForkNode) Close() error {
	n.inPort.Close()
	n.outPort.Close()
	n.errPort.Close()
	return nil
}

func (n *ForkNode) forward(proc *process.Process) {
	inReader := n.inPort.Open(proc)

	for inPck := range inReader.Read() {
		child := proc.Fork()
		outWriter := n.outPort.Open(child)

		outWriter.Write(inPck)
		inReader.Receive(packet.None)
	}
}

func (n *ForkNode) backward(proc *process.Process) {
	outWriter := n.outPort.Open(proc)
	var errWriter *packet.Writer

	for backPck := range outWriter.Receive() {
		var err error
		if v, ok := backPck.Payload().(types.Error); ok {
			err = v.Unwrap()
		}

		if err != nil {
			if errWriter == nil {
				errWriter = n.errPort.Open(proc)
			}
			if errWriter.Write(backPck) > 0 {
				continue
			}
		}

		proc.Join()
		proc.Exit(err)
	}
}

func (n *ForkNode) catch(proc *process.Process) {
	errWriter := n.errPort.Open(proc)

	for backPck := range errWriter.Receive() {
		var err error
		if v, ok := backPck.Payload().(types.Error); ok {
			err = v.Unwrap()
		}

		proc.Join()
		proc.Exit(err)
	}
}
