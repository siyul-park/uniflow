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
	}

	n.inPort.AddListener(port.ListenFunc(n.forward))
	n.outPorts[0].AddListener(port.ListenFunc(n.backward0))
	n.outPorts[1].AddListener(port.ListenFunc(n.backward1))

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
	n.tracer.Close()
	return nil
}

func (n *TestNode) forward(proc *process.Process) {
	inReader := n.inPort.Open(proc)
	outWriter0 := n.outPorts[0].Open(proc)
	outWriter1 := n.outPorts[1].Open(proc)

	for inPck := range inReader.Read() {
		n.tracer.Read(inReader, inPck)
		currentPck := inPck

		n.tracer.AddHook(inPck, packet.HookFunc(func(resultPck *packet.Packet) {
			n.tracer.Transform(inPck, resultPck)
			n.handleResult(resultPck, outWriter0, outWriter1)
			inReader.Receive(currentPck)
		}))

		n.tracer.Write(outWriter0, inPck)
	}
}

func (n *TestNode) handleResult(resultPck *packet.Packet, outWriter0, outWriter1 *packet.Writer) {
	n.tracer.Write(outWriter0, resultPck)

	if n.Out(node.PortWithIndex(node.PortOut, 1)).Links() > 0 {
		n.handleValidation(resultPck.Payload(), outWriter1, outWriter0)
	}
}

func (n *TestNode) handleValidation(payload any, outWriter1, outWriter0 *packet.Writer) {
	if slice, ok := payload.(types.Slice); ok {
		n.validateFrames(slice, outWriter1, outWriter0)
	} else {
		n.validateFrame(payload, -1, outWriter1, outWriter0, nil)
	}
}

func (n *TestNode) validateFrames(slice types.Slice, outWriter1, outWriter0 *packet.Writer) {
	n.validateFrame(slice.Get(0), -1, outWriter1, outWriter0, func() {
		for i := 1; i < slice.Len(); i++ {
			n.validateFrame(slice.Get(i), i-1, outWriter1, outWriter0, nil)
		}
	})
}

func (n *TestNode) validateFrame(value any, index int, outWriter1, outWriter0 *packet.Writer, callback func()) {
	validationValue := n.convertToValue(value)
	validationPck := packet.New(types.NewSlice(validationValue, types.NewInt(index)))
	n.tracer.Write(outWriter1, validationPck)

	n.tracer.AddHook(validationPck, packet.HookFunc(func(validationResult *packet.Packet) {
		if err, ok := validationResult.Payload().(types.Error); ok {
			n.tracer.Write(outWriter0, packet.New(err))
		}
		if callback != nil {
			callback()
		}
	}))
}

func (n *TestNode) convertToValue(v any) types.Value {
	switch val := v.(type) {
	case nil:
		return nil
	case types.Error:
		return val
	case types.Value:
		return val
	default:
		return types.NewString(fmt.Sprintf("%v", val))
	}
}

func (n *TestNode) backward0(proc *process.Process) {
	n.handleBackward(n.outPorts[0].Open(proc))
}

func (n *TestNode) backward1(proc *process.Process) {
	n.handleBackward(n.outPorts[1].Open(proc))
}

func (n *TestNode) handleBackward(outWriter *packet.Writer) {
	for backPck := range outWriter.Receive() {
		n.tracer.Receive(outWriter, backPck)
	}
}
