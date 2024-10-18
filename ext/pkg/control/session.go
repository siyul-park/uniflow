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

// SessionNodeSpec defines the specification for creating a SessionNode.
type SessionNodeSpec struct {
	spec.Meta `map:",inline"`
}

// SessionNode manages session data flow and process interactions through its ports.
type SessionNode struct {
	values  *process.Local[types.Value]
	tracer  *packet.Tracer
	ioPort  *port.InPort
	inPort  *port.InPort
	outPort *port.OutPort
}

const KindSession = "session"

var _ node.Node = (*SessionNode)(nil)

// NewSessionNodeCodec creates a codec for decoding NewSessionNodeCodec.
func NewSessionNodeCodec() scheme.Codec {
	return scheme.CodecWithType(func(_ *SessionNodeSpec) (node.Node, error) {
		return NewSessionNode(), nil
	})
}

// NewSessionNode creates and initializes a new SessionNode.
func NewSessionNode() *SessionNode {
	n := &SessionNode{
		values:  process.NewLocal[types.Value](),
		tracer:  packet.NewTracer(),
		ioPort:  port.NewIn(),
		inPort:  port.NewIn(),
		outPort: port.NewOut(),
	}

	n.ioPort.AddListener(port.ListenFunc(n.session))
	n.inPort.AddListener(port.ListenFunc(n.forward))
	n.outPort.AddListener(port.ListenFunc(n.backward))

	return n
}

// In returns the input port with the specified name.
func (n *SessionNode) In(name string) *port.InPort {
	switch name {
	case node.PortIO:
		return n.ioPort
	case node.PortIn:
		return n.inPort
	default:
		return nil
	}
}

// Out returns the output port with the specified name.
func (n *SessionNode) Out(name string) *port.OutPort {
	switch name {
	case node.PortOut:
		return n.outPort
	default:
		return nil
	}
}

// Close closes all ports and associated resources of the node.
func (n *SessionNode) Close() error {
	n.ioPort.Close()
	n.inPort.Close()
	n.outPort.Close()
	n.values.Close()
	n.tracer.Close()
	return nil
}

func (n *SessionNode) session(proc *process.Process) {
	ioReader := n.ioPort.Open(proc)

	for inPck := range ioReader.Read() {
		n.values.Store(proc, inPck.Payload())
		ioReader.Receive(packet.None)
	}
}

func (n *SessionNode) forward(proc *process.Process) {
	inReader := n.inPort.Open(proc)

	for inPck := range inReader.Read() {
		n.tracer.Read(inReader, inPck)

		parents := n.values.Keys()
		children := make([]*process.Process, 0, len(parents))
		for _, parent := range parents {
			children = append(children, parent.Fork())
		}

		outPcks := make([]*packet.Packet, 0, len(children))
		for i := 0; i < len(children); i++ {
			child := children[i]
			if value, ok := n.values.Load(child); ok {
				outPck := packet.New(types.NewSlice(value, inPck.Payload()))
				n.tracer.Transform(inPck, outPck)
				outPcks = append(outPcks, outPck)
			} else {
				child.Exit(nil)
				children = append(children[:i], children[i+1:]...)
				i--
			}
		}

		n.tracer.AddHook(inPck, packet.HookFunc(func(backPck *packet.Packet) {
			var err error
			if v, ok := backPck.Payload().(types.Error); ok {
				err = v.Unwrap()
			}

			for _, child := range children {
				child.Join()
				child.Exit(err)
			}
		}))

		for i, outPck := range outPcks {
			outWriter := n.outPort.Open(children[i])
			n.tracer.Write(outWriter, outPck)
		}
		if len(outPcks) == 0 {
			n.tracer.Reduce(inPck)
		}
	}
}

func (n *SessionNode) backward(proc *process.Process) {
	outWriter := n.outPort.Open(proc)

	for backPck := range outWriter.Receive() {
		n.tracer.Receive(outWriter, backPck)
	}
}
