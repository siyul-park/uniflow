package control

import (
	"sync"

	"github.com/siyul-park/uniflow/pkg/node"
	"github.com/siyul-park/uniflow/pkg/object"
	"github.com/siyul-park/uniflow/pkg/packet"
	"github.com/siyul-park/uniflow/pkg/process"
	"github.com/siyul-park/uniflow/pkg/spec"
	"github.com/siyul-park/uniflow/plugin/internal/language"
)

// SnippetNode represents a node that executes code snippets in various language.
type SnippetNode struct {
	*node.OneToOneNode
	transform func(any) (any, error)
	mu        sync.RWMutex
}

// SnippetNodeSpec holds the specifications for creating a SnippetNode.
type SnippetNodeSpec struct {
	spec.Meta `map:",inline"`
	Lang      string `map:"lang,omitempty"`
	Code      string `map:"code"`
}

const KindSnippet = "snippet"

// NewSnippetNode creates a new SnippetNode with the specified language.Language and code.
func NewSnippetNode(transform func(any) (any, error)) *SnippetNode {
	n := &SnippetNode{transform: transform}
	n.OneToOneNode = node.NewOneToOneNode(n.action)

	return n
}

func (n *SnippetNode) action(_ *process.Process, inPck *packet.Packet) (*packet.Packet, *packet.Packet) {
	n.mu.RLock()
	defer n.mu.RUnlock()

	inPayload := inPck.Payload()
	input := object.InterfaceOf(inPayload)

	if output, err := n.transform(input); err != nil {
		return nil, packet.New(object.NewError(err))
	} else if outPayload, err := object.MarshalText(output); err != nil {
		return nil, packet.New(object.NewError(err))
	} else {
		return packet.New(outPayload), nil
	}
}

// NewSnippetNodeCodec creates a new codec for SnippetNodeSpec.
func NewSnippetNodeCodec() spec.Codec {
	return spec.CodecWithType(func(spec *SnippetNodeSpec) (node.Node, error) {
		lang := spec.Lang
		transform, err := language.CompileTransform(spec.Code, &lang)
		if err != nil {
			return nil, err
		}

		return NewSnippetNode(transform), nil
	})
}
