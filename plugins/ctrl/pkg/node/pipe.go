package node

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

// PipeNodeSpec holds the specification for creating a PipeNode.
type PipeNodeSpec struct {
	spec.Meta `json:",inline"`
}

// PipeNode processes an input packet and sends the result to multiple output ports.
type PipeNode struct {
	tracer   *packet.Tracer
	inPort   *port.InPort
	outPorts []*port.OutPort
	errPort  *port.OutPort
	mu       sync.RWMutex
}

const KindPipe = "pipe"

var _ node.Node = (*PipeNode)(nil)

// NewPipeNodeCodec creates a new codec for PipeNodeSpec.
func NewPipeNodeCodec() scheme.Codec {
	return scheme.CodecWithType(func(spec *PipeNodeSpec) (node.Node, error) {
		return NewPipeNode(), nil
	})
}

// NewPipeNode creates a new PipeNode.
func NewPipeNode() *PipeNode {
	n := &PipeNode{
		tracer:  packet.NewTracer(),
		inPort:  port.NewIn(),
		errPort: port.NewOut(),
	}

	n.inPort.AddListener(port.ListenFunc(n.forward))
	n.errPort.AddListener(port.ListenFunc(n.catch))

	return n
}

// In returns the input port with the specified name.
func (n *PipeNode) In(name string) *port.InPort {
	switch name {
	case node.PortIn:
		return n.inPort
	default:
		return nil
	}
}

// Out returns the output port with the specified name.
func (n *PipeNode) Out(name string) *port.OutPort {
	n.mu.Lock()
	defer n.mu.Unlock()

	if name == node.PortError {
		return n.errPort
	}
	if node.NameOfPort(name) == node.PortOut {
		index, _ := node.IndexOfPort(name)
		for i := 0; i <= index; i++ {
			if len(n.outPorts) <= i {
				outPort := port.NewOut()
				n.outPorts = append(n.outPorts, outPort)
				outPort.AddListener(n.backward(i))
			}
		}
		return n.outPorts[index]
	}
	return nil
}

// Close closes all ports associated with the node.
func (n *PipeNode) Close() error {
	n.mu.RLock()
	defer n.mu.RUnlock()

	n.inPort.Close()
	for _, outPort := range n.outPorts {
		outPort.Close()
	}
	n.errPort.Close()
	n.tracer.Close()
	return nil
}

func (n *PipeNode) forward(proc *process.Process) {
	n.mu.RLock()
	defer n.mu.RUnlock()

	inReader := n.inPort.Open(proc)
	var outWriter0 *packet.Writer

	for inPck := range inReader.Read() {
		n.tracer.Read(inReader, inPck)

		if outWriter0 == nil && len(n.outPorts) > 0 {
			outWriter0 = n.outPorts[0].Open(proc)
		}
		n.tracer.Write(outWriter0, inPck)
	}
}

func (n *PipeNode) backward(index int) port.Listener {
	return port.ListenFunc(func(proc *process.Process) {
		n.mu.RLock()
		defer n.mu.RUnlock()

		outPort0 := n.outPorts[index]
		var outPort1 *port.OutPort
		for i := index + 1; i < len(n.outPorts); i++ {
			if n.outPorts[i] != nil && len(n.outPorts[i].Links()) > 0 {
				outPort1 = n.outPorts[i]
				break
			}
		}

		outWriter0 := outPort0.Open(proc)
		var outWriter1 *packet.Writer
		var errWriter *packet.Writer

		for backPck := range outWriter0.Receive() {
			var outPck *packet.Packet
			if outPcks := n.tracer.Writes(outWriter0); len(outPcks) > 0 {
				outPck = outPcks[0]
			}

			if _, ok := backPck.Payload().(types.Error); ok {
				if errWriter == nil {
					errWriter = n.errPort.Open(proc)
				}
				n.tracer.Link(outPck, backPck)
				n.tracer.Write(errWriter, backPck)
				n.tracer.Receive(outWriter0, nil)
			} else if outPort1 != nil {
				if outWriter1 == nil {
					outWriter1 = outPort1.Open(proc)
				}
				n.tracer.Link(outPck, backPck)
				n.tracer.Write(outWriter1, backPck)
				n.tracer.Receive(outWriter0, nil)
			} else {
				n.tracer.Receive(outWriter0, backPck)
			}
		}
	})
}

func (n *PipeNode) catch(proc *process.Process) {
	errWriter := n.errPort.Open(proc)

	for backPck := range errWriter.Receive() {
		n.tracer.Receive(errWriter, backPck)
	}
}
