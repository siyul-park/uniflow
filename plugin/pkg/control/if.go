package control

import (
	"reflect"
	"sync"

	"github.com/siyul-park/uniflow/pkg/node"
	"github.com/siyul-park/uniflow/pkg/object"
	"github.com/siyul-park/uniflow/pkg/packet"
	"github.com/siyul-park/uniflow/pkg/port"
	"github.com/siyul-park/uniflow/pkg/process"
	"github.com/siyul-park/uniflow/pkg/scheme"
	"github.com/siyul-park/uniflow/pkg/spec"
	"github.com/siyul-park/uniflow/plugin/pkg/language"
)

// IfNode represents a node that evaluates a condition and forwards packets based on the result.
type IfNode struct {
	when     func(any) (bool, error)
	tracers  *process.Local[*packet.Tracer]
	inPort   *port.InPort
	outPorts []*port.OutPort
	errPort  *port.OutPort
	mu       sync.RWMutex
}

// IfNodeSpec holds the specifications for creating a IfNode.
type IfNodeSpec struct {
	spec.Meta `map:",inline"`
	When      string `map:"when"`
}

var _ node.Node = (*IfNode)(nil)

const KindIf = "if"

// NewIfNode creates a new IfNode.
func NewIfNode(when func(any) (bool, error)) *IfNode {
	n := &IfNode{
		when:     when,
		tracers:  process.NewLocal[*packet.Tracer](),
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
	n.tracers.Close()

	return nil
}

func (n *IfNode) forward(proc *process.Process) {
	n.mu.RLock()
	defer n.mu.RUnlock()

	tracer, _ := n.tracers.LoadOrStore(proc, func() (*packet.Tracer, error) {
		return packet.NewTracer(), nil
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
		tracer.Read(inReader, inPck)

		inPayload := inPck.Payload()
		input := object.InterfaceOf(inPayload)

		if ok, err := n.when(input); err != nil {
			errPck := packet.New(object.NewError(err))
			tracer.Transform(inPck, errPck)
			tracer.Write(errWriter, errPck)
		} else if ok {
			tracer.Write(outWriter0, inPck)
		} else {
			tracer.Write(outWriter1, inPck)
		}
	}
}

func (n *IfNode) backward(proc *process.Process, index int) {
	n.mu.RLock()
	defer n.mu.RUnlock()

	tracer, _ := n.tracers.LoadOrStore(proc, func() (*packet.Tracer, error) {
		return packet.NewTracer(), nil
	})

	outWriter := n.outPorts[index].Open(proc)

	for {
		backPck, ok := <-outWriter.Receive()
		if !ok {
			return
		}

		tracer.Receive(outWriter, backPck)
	}
}

func (n *IfNode) catch(proc *process.Process) {
	n.mu.RLock()
	defer n.mu.RUnlock()

	tracer, _ := n.tracers.LoadOrStore(proc, func() (*packet.Tracer, error) {
		return packet.NewTracer(), nil
	})

	errWriter := n.errPort.Open(proc)

	for {
		backPck, ok := <-errWriter.Receive()
		if !ok {
			return
		}

		tracer.Receive(errWriter, backPck)
	}
}

// NewIfNodeCodec creates a new codec for IfNodeSpec.
func NewIfNodeCodec(compiler language.Compiler) scheme.Codec {
	return scheme.CodecWithType(func(spec *IfNodeSpec) (node.Node, error) {
		program, err := compiler.Compile(spec.When)
		if err != nil {
			return nil, err
		}
		return NewIfNode(func(env any) (bool, error) {
			res, err := program.Run(env)
			if err != nil {
				return false, err
			}
			return !reflect.ValueOf(res).IsZero(), nil
		}), nil
	})
}
