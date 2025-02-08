package control

import (
	"fmt"

	"github.com/pkg/errors"
	"github.com/siyul-park/uniflow/pkg/node"
	"github.com/siyul-park/uniflow/pkg/packet"
	"github.com/siyul-park/uniflow/pkg/port"
	"github.com/siyul-park/uniflow/pkg/process"
	"github.com/siyul-park/uniflow/pkg/scheme"
	"github.com/siyul-park/uniflow/pkg/spec"
	"github.com/siyul-park/uniflow/pkg/types"
)

// ThrowNodeSpec defines the specification for the ThrowNode.
type ThrowNodeSpec struct {
	spec.Meta `json:",inline"`
}

// ThrowNode represents a node that throws errors based on incoming packets.
type ThrowNode struct {
	inPort *port.InPort
}

const KindThrow = "throw"

var _ node.Node = (*ThrowNode)(nil)

// NewThrowNodeCodec creates a codec for decoding ThrowNodeSpec.
func NewThrowNodeCodec() scheme.Codec {
	return scheme.CodecWithType(func(_ *ThrowNodeSpec) (node.Node, error) {
		return NewThrowNode(), nil
	})
}

// NewThrowNode creates a new ThrowNode instance.
func NewThrowNode() *ThrowNode {
	n := &ThrowNode{
		inPort: port.NewIn(),
	}
	n.inPort.AddListener(port.ListenFunc(n.forward))
	return n
}

// In returns the input port for the given name.
func (n *ThrowNode) In(name string) *port.InPort {
	switch name {
	case node.PortIn:
		return n.inPort
	default:
		return nil
	}
}

// Out returns the output port for the given name.
func (n *ThrowNode) Out(_ string) *port.OutPort {
	return nil
}

// Close closes the input port.
func (n *ThrowNode) Close() error {
	n.inPort.Close()
	return nil
}

func (n *ThrowNode) forward(proc *process.Process) {
	inReader := n.inPort.Open(proc)

	for inPck := range inReader.Read() {
		inPayload := inPck.Payload()

		var messages []string
		if vals, ok := inPayload.(types.Slice); ok {
			for _, val := range vals.Range() {
				if val != nil {
					messages = append(messages, fmt.Sprint(types.InterfaceOf(val)))
				}
			}
		} else if inPayload != nil {
			messages = append(messages, fmt.Sprint(types.InterfaceOf(inPayload)))
		}

		var err error
		for _, message := range messages {
			if err == nil {
				err = errors.New(message)
			} else {
				err = errors.WithMessage(err, message)
			}
		}

		var outPayload types.Value
		if err != nil {
			outPayload = types.NewError(err)
		}

		inReader.Receive(packet.New(outPayload))
	}
}
