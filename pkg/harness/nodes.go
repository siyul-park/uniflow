package harness

import (
	"log"

	"github.com/siyul-park/uniflow/pkg/node"
	"github.com/siyul-park/uniflow/pkg/packet"
	"github.com/siyul-park/uniflow/pkg/process"
	"github.com/siyul-park/uniflow/pkg/types"
)

// SourceNode generates data for testing
type SourceNode struct {
	*node.OneToOneNode
	data types.Value
}

// NewSourceNode creates a new source node with the given data
func NewSourceNode(data types.Value) *SourceNode {
	n := &SourceNode{
		data: data,
	}
	n.OneToOneNode = node.NewOneToOneNode(n.action)
	return n
}

func (n *SourceNode) action(proc *process.Process, in *packet.Packet) (*packet.Packet, *packet.Packet) {
	log.Printf("SourceNode received input: %v", in)
	// Always output data when triggered, regardless of input content
	if in != nil && n.data != nil {
		out := packet.New(n.data)
		log.Printf("SourceNode producing output: %v", out.Payload())
		return out, packet.None
	}
	log.Printf("SourceNode skipping (no input or no data)")
	return packet.None, packet.None
}

// TransformNode performs operations on input data
type TransformNode struct {
	*node.OneToOneNode
	transform func(types.Value) (types.Value, error)
}

// NewTransformNode creates a new transform node with the given transform function
func NewTransformNode(transform func(types.Value) (types.Value, error)) *TransformNode {
	n := &TransformNode{
		transform: transform,
	}
	n.OneToOneNode = node.NewOneToOneNode(n.action)
	return n
}

func (n *TransformNode) action(proc *process.Process, in *packet.Packet) (*packet.Packet, *packet.Packet) {
	log.Printf("TransformNode received input: %v", in)
	if in == nil || n.transform == nil {
		log.Printf("TransformNode skipping (no input or no transform)")
		return packet.None, packet.None
	}

	result, err := n.transform(in.Payload())
	if err != nil {
		log.Printf("TransformNode error: %v", err)
		return packet.None, packet.New(types.NewError(err))
	}
	if result == nil {
		log.Printf("TransformNode skipping (nil result)")
		return packet.None, packet.None
	}
	log.Printf("TransformNode producing output: %v", result)
	return packet.New(result), packet.None
}

// AssertNode validates test results
type AssertNode struct {
	*node.OneToOneNode
	resultChan chan<- *packet.Packet
	assert     func(types.Value) error
}

// NewAssertNode creates a new assert node
func NewAssertNode(resultChan chan<- *packet.Packet, assert func(types.Value) error) *AssertNode {
	n := &AssertNode{
		resultChan: resultChan,
		assert:     assert,
	}
	n.OneToOneNode = node.NewOneToOneNode(n.action)
	return n
}

func (n *AssertNode) action(proc *process.Process, in *packet.Packet) (*packet.Packet, *packet.Packet) {
	log.Printf("AssertNode received input: %v", in)
	if in == nil {
		log.Printf("AssertNode skipping (no input)")
		return packet.None, packet.None
	}

	if n.assert != nil {
		if err := n.assert(in.Payload()); err != nil {
			log.Printf("AssertNode assertion failed: %v", err)
			return packet.None, packet.New(types.NewError(err))
		}
	}

	if n.resultChan != nil {
		log.Printf("AssertNode sending to result channel: %v", in.Payload())
		n.resultChan <- in
	}
	return packet.None, packet.None
}
