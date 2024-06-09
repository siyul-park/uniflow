package control

import (
	"reflect"
	"sync"

	"github.com/siyul-park/uniflow/pkg/node"
	"github.com/siyul-park/uniflow/pkg/object"
	"github.com/siyul-park/uniflow/pkg/packet"
	"github.com/siyul-park/uniflow/pkg/port"
	"github.com/siyul-park/uniflow/pkg/process"
	"github.com/siyul-park/uniflow/pkg/spec"
	"github.com/siyul-park/uniflow/plugin/internal/language"
)

// IfNode represents a node that evaluates a condition and forwards packets based on the result.
type IfNode struct {
	when     func(any) (bool, error)
	bridges  *process.Local[*packet.Bridge]
	inPort   *port.InPort
	outPorts []*port.OutPort
	errPort  *port.OutPort
	mu       sync.RWMutex
}

// IfNodeSpec holds the specifications for creating a IfNode.
type IfNodeSpec struct {
	spec.Meta `map:",inline"`
	Lang      string `map:"lang,omitempty"`
	When      string `map:"when"`
}

var _ node.Node = (*IfNode)(nil)

const KindIf = "if"

// NewIfNode creates a new IfNode.
func NewIfNode(when func(any) (bool, error)) *IfNode {
	n := &IfNode{
		when:     when,
		bridges:  process.NewLocal[*packet.Bridge](),
		inPort:   port.NewIn(),
		outPorts: []*port.OutPort{port.NewOut(), port.NewOut()},
		errPort:  port.NewOut(),
	}

	n.inPort.AddInitHook(port.InitHookFunc(n.forward))
	for i, outPort := range n.outPorts {
		i := i
		outPort.AddInitHook(port.InitHookFunc(func(proc *process.Process) {
			n.backward(proc, i)
		}))
	}
	n.errPort.AddInitHook(port.InitHookFunc(n.catch))

	return n
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
	n.bridges.Close()

	return nil
}

func (n *IfNode) forward(proc *process.Process) {
	n.mu.RLock()
	defer n.mu.RUnlock()

	bridge, _ := n.bridges.LoadOrStore(proc, func() (*packet.Bridge, error) {
		return packet.NewBridge(), nil
	})

	inReader := n.inPort.Open(proc)
	outWriter0 := n.outPorts[0].Open(proc)
	outWriter1 := n.outPorts[1].Open(proc)
	errWriter := n.errPort.Open(proc)

	for {
		inPck, ok := <-inReader.Read()
		if !ok {
			return
		}

		inPayload := inPck.Payload()
		input := object.InterfaceOf(inPayload)

		if ok, err := n.when(input); err != nil {
			errPck := packet.New(object.NewError(err))
			bridge.Write([]*packet.Packet{errPck}, []*packet.Reader{inReader}, []*packet.Writer{errWriter})
		} else if ok {
			bridge.Write([]*packet.Packet{inPck}, []*packet.Reader{inReader}, []*packet.Writer{outWriter0})
		} else {
			bridge.Write([]*packet.Packet{inPck}, []*packet.Reader{inReader}, []*packet.Writer{outWriter1})
		}
	}
}

func (n *IfNode) backward(proc *process.Process, index int) {
	n.mu.RLock()
	defer n.mu.RUnlock()

	bridge, _ := n.bridges.LoadOrStore(proc, func() (*packet.Bridge, error) {
		return packet.NewBridge(), nil
	})

	outWriter := n.outPorts[index].Open(proc)

	for {
		backPck, ok := <-outWriter.Receive()
		if !ok {
			return
		}

		bridge.Receive(backPck, outWriter)
	}
}

func (n *IfNode) catch(proc *process.Process) {
	n.mu.RLock()
	defer n.mu.RUnlock()

	bridge, _ := n.bridges.LoadOrStore(proc, func() (*packet.Bridge, error) {
		return packet.NewBridge(), nil
	})

	errWriter := n.errPort.Open(proc)

	for {
		backPck, ok := <-errWriter.Receive()
		if !ok {
			return
		}

		bridge.Receive(backPck, errWriter)
	}
}

// NewIfNodeCodec creates a new codec for IfNodeSpec.
func NewIfNodeCodec() spec.Codec {
	return spec.CodecWithType(func(spec *IfNodeSpec) (node.Node, error) {
		lang := spec.Lang
		transform, err := language.CompileTransform(spec.When, &lang)
		if err != nil {
			return nil, err
		}

		return NewIfNode(func(input any) (bool, error) {
			output, err := transform(input)
			if err != nil {
				return false, err
			}
			return !reflect.ValueOf(output).IsZero(), nil
		}), nil
	})
}
