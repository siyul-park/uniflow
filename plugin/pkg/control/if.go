package control

import (
	"reflect"
	"sync"

	"github.com/siyul-park/uniflow/pkg/node"
	"github.com/siyul-park/uniflow/pkg/packet"
	"github.com/siyul-park/uniflow/pkg/port"
	"github.com/siyul-park/uniflow/pkg/primitive"
	"github.com/siyul-park/uniflow/pkg/process"
	"github.com/siyul-park/uniflow/pkg/scheme"
	"github.com/siyul-park/uniflow/plugin/internal/language"
)

// IfNode represents a node that evaluates a condition and forwards packets based on the result.
type IfNode struct {
	lang     string
	when     func(primitive.Value) (bool, error)
	inPort   *port.InPort
	outPorts []*port.OutPort
	errPort  *port.OutPort
	mu       sync.RWMutex
}

// IfNodeSpec holds the specifications for creating a IfNode.
type IfNodeSpec struct {
	scheme.SpecMeta `map:",inline"`
	Lang            string `map:"lang,omitempty"`
	When            string `map:"when"`
}

var _ node.Node = (*IfNode)(nil)

const KindIf = "if"

// NewIfNode creates a new IfNode.
func NewIfNode(code, lang string) (*IfNode, error) {
	if lang == "" {
		lang = language.Detect(code)
	}

	transform, err := language.CompileTransformWithPrimitive(code, lang)
	if err != nil {
		return nil, err
	}

	n := &IfNode{
		lang: lang,
		when: func(value primitive.Value) (bool, error) {
			output, err := transform(value)
			if err != nil {
				return false, err
			}
			return !reflect.ValueOf(output.Interface()).IsZero(), nil
		},
		inPort:   port.NewIn(),
		outPorts: []*port.OutPort{port.NewOut(), port.NewOut()},
		errPort:  port.NewOut(),
	}

	n.inPort.AddHandler(port.HandlerFunc(n.forward))
	for i, outPort := range n.outPorts {
		i := i
		outPort.AddHandler(port.HandlerFunc(func(proc *process.Process) {
			n.backward(proc, i)
		}))
	}
	n.errPort.AddHandler(port.HandlerFunc(n.catch))

	return n, nil
}

// In returns the input port with the specified name.
func (n *IfNode) In(name string) *port.InPort {
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
func (n *IfNode) Out(name string) *port.OutPort {
	n.mu.RLock()
	defer n.mu.RUnlock()

	switch name {
	case node.PortOut:
		return n.outPorts[0]
	case node.PortErr:
		return n.errPort
	default:
		if i, ok := node.IndexOfPort(node.PortOut, name); ok {
			if len(n.outPorts) > i {
				return n.outPorts[i]
			}
		}
	}

	return nil
}

// Close closes all ports associated with the node.
func (n *IfNode) Close() error {
	n.mu.Lock()
	defer n.mu.Unlock()

	n.inPort.Close()
	for _, p := range n.outPorts {
		p.Close()
	}
	n.errPort.Close()

	return nil
}

func (n *IfNode) forward(proc *process.Process) {
	n.mu.RLock()
	defer n.mu.RUnlock()

	inReader := n.inPort.Open(proc)
	outWriter0 := n.outPorts[0].Open(proc)
	outWriter1 := n.outPorts[1].Open(proc)

	for {
		inPck, ok := <-inReader.Read()
		if !ok {
			return
		}

		if ok, err := n.when(inPck.Payload()); err != nil {
			errPck := packet.WithError(err, inPck)
			proc.Stack().Add(inPck, errPck)
			n.throw(proc, errPck)
		} else if ok {
			if !outWriter0.Write(inPck) {
				if !inReader.Receive(inPck) {
					proc.Stack().Clear(inPck)
				}
			}
		} else {
			if !outWriter1.Write(inPck) {
				if !inReader.Receive(inPck) {
					proc.Stack().Clear(inPck)
				}
			}
		}
	}
}

func (n *IfNode) backward(proc *process.Process, index int) {
	n.mu.RLock()
	defer n.mu.RUnlock()

	inReader := n.inPort.Open(proc)
	outWriter := n.outPorts[index].Open(proc)

	for {
		backPck, ok := <-outWriter.Receive()
		if !ok {
			return
		}

		if !inReader.Receive(backPck) {
			proc.Stack().Clear(backPck)
		}
	}
}

func (n *IfNode) catch(proc *process.Process) {
	n.mu.RLock()
	defer n.mu.RUnlock()

	inReader := n.inPort.Open(proc)
	errWriter := n.errPort.Open(proc)

	for {
		backPck, ok := <-errWriter.Receive()
		if !ok {
			return
		}

		if !inReader.Receive(backPck) {
			proc.Stack().Clear(backPck)
		}
	}
}

func (n *IfNode) throw(proc *process.Process, errPck *packet.Packet) {
	inReader := n.inPort.Open(proc)
	errWriter := n.errPort.Open(proc)

	if !errWriter.Write(errPck) {
		if !inReader.Receive(errPck) {
			proc.Stack().Clear(errPck)
		}
	}
}

// NewIfNodeCodec creates a new codec for IfNodeSpec.
func NewIfNodeCodec() scheme.Codec {
	return scheme.CodecWithType(func(spec *IfNodeSpec) (node.Node, error) {
		return NewIfNode(spec.When, spec.Lang)
	})
}
