package control

import (
	"sync"

	"github.com/siyul-park/uniflow/ext/pkg/language"
	"github.com/siyul-park/uniflow/pkg/node"
	"github.com/siyul-park/uniflow/pkg/object"
	"github.com/siyul-park/uniflow/pkg/packet"
	"github.com/siyul-park/uniflow/pkg/process"
	"github.com/siyul-park/uniflow/pkg/scheme"
	"github.com/siyul-park/uniflow/pkg/spec"
)

// SnippetNode represents a node that executes code snippets in various language.
type SnippetNode struct {
	*node.OneToOneNode
	fn func(any) (any, error)
	mu sync.RWMutex
}

// SnippetNodeSpec holds the specifications for creating a SnippetNode.
type SnippetNodeSpec struct {
	spec.Meta `map:",inline"`
	Language  string `map:"language,omitempty"`
	Code      string `map:"code"`
}

const KindSnippet = "snippet"

// NewSnippetNode creates a new SnippetNode with the specified language.Language and code.
func NewSnippetNode(fn func(any) (any, error)) *SnippetNode {
	n := &SnippetNode{fn: fn}
	n.OneToOneNode = node.NewOneToOneNode(n.action)

	return n
}

func (n *SnippetNode) action(_ *process.Process, inPck *packet.Packet) (*packet.Packet, *packet.Packet) {
	n.mu.RLock()
	defer n.mu.RUnlock()

	inPayload := inPck.Payload()
	input := object.InterfaceOf(inPayload)

	if output, err := n.fn(input); err != nil {
		return nil, packet.New(object.NewError(err))
	} else if outPayload, err := object.MarshalText(output); err != nil {
		return nil, packet.New(object.NewError(err))
	} else {
		return packet.New(outPayload), nil
	}
}

// NewSnippetNodeCodec creates a new codec for SnippetNodeSpec.
func NewSnippetNodeCodec(module *language.Module) scheme.Codec {
	return scheme.CodecWithType(func(spec *SnippetNodeSpec) (node.Node, error) {
		compiler, err := module.Load(spec.Language)
		if err != nil {
			return nil, err
		}

		program, err := compiler.Compile(spec.Code)
		if err != nil {
			return nil, err
		}

		return NewSnippetNode(program.Run), nil
	})
}
