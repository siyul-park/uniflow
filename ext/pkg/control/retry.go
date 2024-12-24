package control

import (
	"github.com/siyul-park/uniflow/pkg/node"
	"github.com/siyul-park/uniflow/pkg/packet"
	"github.com/siyul-park/uniflow/pkg/port"
	"github.com/siyul-park/uniflow/pkg/process"
	"github.com/siyul-park/uniflow/pkg/scheme"
	"github.com/siyul-park/uniflow/pkg/spec"
	"github.com/siyul-park/uniflow/pkg/types"
	"sync"
)

// RetryNodeSpec defines the configuration for RetryNode.
type RetryNodeSpec struct {
	spec.Meta `map:",inline"`
	Threshold int `map:"threshold,omitempty"`
}

// RetryNode attempts to process packets up to a specified retry limit.
type RetryNode struct {
	threshold int
	tracer    *packet.Tracer
	inPort    *port.InPort
	outPort   *port.OutPort
	errPort   *port.OutPort
}

var _ node.Node = (*RetryNode)(nil)

const KindRetry = "retry"

// NewRetryNodeCodec creates a codec to build RetryNode from RetryNodeSpec.
func NewRetryNodeCodec() scheme.Codec {
	return scheme.CodecWithType(func(spec *RetryNodeSpec) (node.Node, error) {
		return NewRetryNode(spec.Threshold), nil
	})
}

// NewRetryNode initializes a RetryNode with the given retry limit.
func NewRetryNode(threshold int) *RetryNode {
	n := &RetryNode{
		threshold: threshold,
		tracer:    packet.NewTracer(),
		inPort:    port.NewIn(),
		outPort:   port.NewOut(),
		errPort:   port.NewOut(),
	}

	n.inPort.AddListener(port.ListenFunc(n.forward))
	n.outPort.AddListener(port.ListenFunc(n.backward))
	n.errPort.AddListener(port.ListenFunc(n.catch))

	return n
}

// In returns the input port based on the given name.
func (n *RetryNode) In(name string) *port.InPort {
	if name == node.PortIn {
		return n.inPort
	}
	return nil
}

// Out returns the output port based on the given name.
func (n *RetryNode) Out(name string) *port.OutPort {
	switch name {
	case node.PortOut:
		return n.outPort
	case node.PortError:
		return n.errPort
	default:
		return nil
	}
}

// Close shuts down all ports and releases resources.
func (n *RetryNode) Close() error {
	n.inPort.Close()
	n.outPort.Close()
	n.errPort.Close()
	n.tracer.Close()
	return nil
}

func (n *RetryNode) forward(proc *process.Process) {
	inReader := n.inPort.Open(proc)
	outWriter := n.outPort.Open(proc)
	errWriter := n.errPort.Open(proc)

	attempts := &sync.Map{}

	for inPck := range inReader.Read() {
		n.tracer.Read(inReader, inPck)

		var hook packet.Hook
		hook = packet.HookFunc(func(backPck *packet.Packet) {
			if _, ok := backPck.Payload().(types.Error); !ok {
				n.tracer.Transform(inPck, backPck)
				n.tracer.Reduce(backPck)
				return
			}

			for {
				actual, _ := attempts.LoadOrStore(inPck, 0)
				count := actual.(int)

				if count == n.threshold {
					n.tracer.Transform(inPck, backPck)
					n.tracer.Write(errWriter, backPck)
					return
				}

				if attempts.CompareAndSwap(inPck, count, count+1) {
					break
				}
			}

			n.tracer.AddHook(inPck, hook)
			n.tracer.Write(outWriter, inPck)
		})

		n.tracer.AddHook(inPck, hook)
		n.tracer.Write(outWriter, inPck)
	}
}

func (n *RetryNode) backward(proc *process.Process) {
	outWriter := n.outPort.Open(proc)

	for backPck := range outWriter.Receive() {
		n.tracer.Receive(outWriter, backPck)
	}
}

func (n *RetryNode) catch(proc *process.Process) {
	errWriter := n.errPort.Open(proc)

	for backPck := range errWriter.Receive() {
		n.tracer.Receive(errWriter, backPck)
	}
}
