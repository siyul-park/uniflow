package testing

import (
	"fmt"

	"github.com/siyul-park/uniflow/pkg/node"
	"github.com/siyul-park/uniflow/pkg/packet"
	"github.com/siyul-park/uniflow/pkg/port"
	"github.com/siyul-park/uniflow/pkg/process"
	"github.com/siyul-park/uniflow/pkg/scheme"
	"github.com/siyul-park/uniflow/pkg/spec"
	"github.com/siyul-park/uniflow/pkg/types"
)

type TestNodeSpec struct {
	spec.Meta `json:",inline"`
}

type TestNode struct {
	tracer   *packet.Tracer
	inPort   *port.InPort
	outPorts []*port.OutPort
	errPort  *port.OutPort
}

const KindTest = "test"

var _ node.Node = (*TestNode)(nil)

func NewTestNodeCodec() scheme.Codec {
	return scheme.CodecWithType(func(spec *TestNodeSpec) (node.Node, error) {
		return NewTestNode(), nil
	})
}

func NewTestNode() *TestNode {
	n := &TestNode{
		tracer:   packet.NewTracer(),
		inPort:   port.NewIn(),
		outPorts: []*port.OutPort{port.NewOut(), port.NewOut()},
		errPort:  port.NewOut(),
	}

	n.inPort.AddListener(port.ListenFunc(n.forward))
	n.outPorts[0].AddListener(port.ListenFunc(n.backward0))
	n.outPorts[1].AddListener(port.ListenFunc(n.backward1))
	n.errPort.AddListener(port.ListenFunc(n.catch))

	return n
}

func (n *TestNode) In(name string) *port.InPort {
	switch name {
	case node.PortIn:
		return n.inPort
	default:
		return nil
	}
}

func (n *TestNode) Out(name string) *port.OutPort {
	switch name {
	case node.PortOut:
		return n.outPorts[0]
	case node.PortError:
		return n.errPort
	default:
		if node.NameOfPort(name) == node.PortOut {
			index, ok := node.IndexOfPort(name)
			if ok && index < len(n.outPorts) {
				return n.outPorts[index]
			}
		}
	}
	return nil
}

func (n *TestNode) Close() error {
	n.inPort.Close()
	for _, outPort := range n.outPorts {
		outPort.Close()
	}
	n.errPort.Close()
	n.tracer.Close()
	return nil
}

func (n *TestNode) forward(proc *process.Process) {
	inReader := n.inPort.Open(proc)
	outWriter0 := n.outPorts[0].Open(proc)
	outWriter1 := n.outPorts[1].Open(proc)
	errWriter := n.errPort.Open(proc)

	for inPck := range inReader.Read() {
		n.tracer.Read(inReader, inPck)
		currentPck := inPck

		n.tracer.AddHook(inPck, packet.HookFunc(func(backPck *packet.Packet) {
			n.tracer.Transform(inPck, backPck)

			if err, ok := backPck.Payload().(types.Error); ok {
				n.writeError(errWriter, err)
				inReader.Receive(currentPck)
				return
			}

			n.handleValidation(backPck.Payload(), outWriter1, errWriter)
			inReader.Receive(currentPck)
		}))

		n.tracer.Write(outWriter0, inPck)
	}
}

func (n *TestNode) handleValidation(payload any, outWriter1, errWriter *packet.Writer) {
	if slice, ok := payload.(types.Slice); ok {
		n.validateFrame(slice.Get(0), -1, outWriter1, errWriter, func() {
			for i := 1; i < slice.Len(); i++ {
				n.validateFrame(slice.Get(i), i-1, outWriter1, errWriter, nil)
			}
		})
	} else {
		n.validateFrame(payload, -1, outWriter1, errWriter, nil)
	}
}

func (n *TestNode) validateFrame(value any, index int, outWriter1, errWriter *packet.Writer, callback func()) {
	var validationValue types.Value
	switch v := value.(type) {
	case nil:
		validationValue = nil
	case types.Value:
		validationValue = v
	default:
		n.writeError(errWriter, types.NewError(fmt.Errorf("invalid value type")))
		return
	}

	validationPck := packet.New(types.NewSlice(validationValue, types.NewInt(index)))
	n.tracer.Write(outWriter1, validationPck)

	n.tracer.AddHook(validationPck, packet.HookFunc(func(validationResult *packet.Packet) {
		if err, ok := validationResult.Payload().(types.Error); ok {
			n.writeError(errWriter, err)
		}
		if callback != nil {
			callback()
		}
	}))
}

func (n *TestNode) writeError(errWriter *packet.Writer, err types.Error) {
	n.tracer.Write(errWriter, packet.New(err))
}

func (n *TestNode) backward0(proc *process.Process) {
	outWriter := n.outPorts[0].Open(proc)
	for backPck := range outWriter.Receive() {
		n.tracer.Receive(outWriter, backPck)
	}
}

func (n *TestNode) backward1(proc *process.Process) {
	outWriter := n.outPorts[1].Open(proc)
	for backPck := range outWriter.Receive() {
		n.tracer.Receive(outWriter, backPck)
	}
}

func (n *TestNode) catch(proc *process.Process) {
	errWriter := n.errPort.Open(proc)
	for backPck := range errWriter.Receive() {
		n.tracer.Receive(errWriter, backPck)
	}
}
