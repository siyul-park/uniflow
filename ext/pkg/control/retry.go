package control

import (
	"sync"

	"github.com/siyul-park/uniflow/pkg/node"
	"github.com/siyul-park/uniflow/pkg/packet"
	"github.com/siyul-park/uniflow/pkg/port"
	"github.com/siyul-park/uniflow/pkg/process"
	"github.com/siyul-park/uniflow/pkg/scheme"
	"github.com/siyul-park/uniflow/pkg/spec"
	"github.com/siyul-park/uniflow/pkg/types"
)

// RetryNodeSpec defines the configuration for RetryNode.
type RetryNodeSpec struct {
	spec.Meta `json:",inline"`
	Threshold int `json:"threshold,omitempty"`
}

// RetryNode attempts to process packets up to a specified retry limit.
type RetryNode struct {
	threshold int
	tracer    *packet.Tracer
	inPort    *port.InPort
	outPort   *port.OutPort
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
	}

	n.inPort.AddListener(port.ListenFunc(n.forward))
	n.outPort.AddListener(port.ListenFunc(n.backward))

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
	default:
		return nil
	}
}

// Close shuts down all ports and releases resources.
func (n *RetryNode) Close() error {
	n.inPort.Close()
	n.outPort.Close()
	n.tracer.Close()
	return nil
}

func (n *RetryNode) forward(proc *process.Process) {
	inReader := n.inPort.Open(proc)
	var outWriter *packet.Writer

	attempts := &sync.Map{}

	for inPck := range inReader.Read() {
		n.tracer.Read(inReader, inPck)

		var hook packet.Hook
		hook = packet.HookFunc(func(backPck *packet.Packet) {
			for {
				actual, _ := attempts.LoadOrStore(inPck, 0)
				count := actual.(int)

				_, fail := backPck.Payload().(types.Error)
				if !fail || count == n.threshold {
					attempts.Delete(inPck)
					n.tracer.Transform(inPck, backPck)
					n.tracer.Reduce(backPck)
					return
				}

				if attempts.CompareAndSwap(inPck, count, count+1) {
					break
				}
			}

			if outWriter == nil {
				outWriter = n.outPort.Open(proc)
			}
			n.tracer.AddHook(inPck, hook)
			n.tracer.Write(outWriter, inPck)
		})

		if outWriter == nil {
			outWriter = n.outPort.Open(proc)
		}
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
