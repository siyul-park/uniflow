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

// SessionNode manages session data flow and process interactions through its ports.
type SessionNode struct {
	values  *process.Local[object.Object]
	tracer  *packet.Tracer
	ioPort  *port.InPort
	inPort  *port.InPort
	outPort *port.OutPort
	mu      sync.RWMutex
}

// SessionNodeSpec defines the specification for creating a NOPNode.
type SessionNodeSpec struct {
	spec.Meta `map:",inline"`
}

const KindSession = "session"

var _ node.Node = (*SessionNode)(nil)

// NewSessionNode creates and initializes a new SessionNode.
func NewSessionNode() *SessionNode {
	n := &SessionNode{
		values:  process.NewLocal[object.Object](),
		tracer:  packet.NewTracer(),
		ioPort:  port.NewIn(),
		inPort:  port.NewIn(),
		outPort: port.NewOut(),
	}

	n.ioPort.AddInitHook(port.InitHookFunc(n.session))
	n.inPort.AddInitHook(port.InitHookFunc(n.forward))
	n.outPort.AddInitHook(port.InitHookFunc(n.backward))

	return n
}

// In returns the input port with the specified name.
func (n *SessionNode) In(name string) *port.InPort {
	n.mu.RLock()
	defer n.mu.RUnlock()

	switch name {
	case node.PortIO:
		return n.ioPort
	case node.PortIn:
		return n.inPort
	default:
	}

	return nil
}

// Out returns the output port with the specified name.
func (n *SessionNode) Out(name string) *port.OutPort {
	n.mu.RLock()
	defer n.mu.RUnlock()

	switch name {
	case node.PortOut:
		return n.outPort
	default:
	}

	return nil
}

// Close closes all ports and associated resources of the node.
func (n *SessionNode) Close() error {
	n.mu.Lock()
	defer n.mu.Unlock()

	n.ioPort.Close()
	n.inPort.Close()
	n.outPort.Close()
	n.values.Close()
	n.tracer.Close()

	return nil
}

func (n *SessionNode) session(proc *process.Process) {
	n.mu.RLock()
	defer n.mu.RUnlock()

	ioReader := n.ioPort.Open(proc)

	for {
		inPck, ok := <-ioReader.Read()
		if !ok {
			return
		}

		n.values.Store(proc, inPck.Payload())
		ioReader.Receive(packet.None)
	}
}

func (n *SessionNode) forward(proc *process.Process) {
	n.mu.RLock()
	defer n.mu.RUnlock()

	inReader := n.inPort.Open(proc)

	for {
		inPck, ok := <-inReader.Read()
		if !ok {
			return
		}
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
				outPck := packet.New(object.NewSlice(value, inPck.Payload()))
				n.tracer.Transform(inPck, outPck)
				outPcks = append(outPcks, outPck)
			} else {
				child.Exit(nil)
				children = append(children[:i], children[i+1:]...)
				i--
			}
		}

		n.tracer.Sniff(inPck, packet.HandlerFunc(func(backPck *packet.Packet) {
			var err error
			if v, ok := backPck.Payload().(object.Error); ok {
				err, _ = v.Interface().(error)
			}

			for _, child := range children {
				child.Wait()
				child.Exit(err)
			}
		}))

		for i, outPck := range outPcks {
			outWriter := n.outPort.Open(children[i])
			n.tracer.Write(outWriter, outPck)
		}
		if len(outPcks) == 0 {
			n.tracer.Transform(inPck, packet.None)
		}
	}
}

func (n *SessionNode) backward(proc *process.Process) {
	n.mu.RLock()
	defer n.mu.RUnlock()

	outWriter := n.outPort.Open(proc)

	for {
		backPck, ok := <-outWriter.Receive()
		if !ok {
			return
		}

		n.tracer.Receive(outWriter, backPck)
	}
}

// NewSessionNodeCodec creates a codec for decoding NewSessionNodeCodec.
func NewSessionNodeCodec() scheme.Codec {
	return scheme.CodecWithType(func(_ *SessionNodeSpec) (node.Node, error) {
		return NewSessionNode(), nil
	})
}
