package control

import (
	"sync"

	"github.com/siyul-park/uniflow/pkg/primitive"

	"github.com/siyul-park/uniflow/pkg/node"
	"github.com/siyul-park/uniflow/pkg/packet"
	"github.com/siyul-park/uniflow/pkg/process"
	"github.com/siyul-park/uniflow/pkg/scheme"
	"github.com/siyul-park/uniflow/plugin/internal/language"
)

// SnippetNode represents a node that executes code snippets in various language.
type SnippetNode struct {
	*node.OneToOneNode
	program func(primitive.Value) (primitive.Value, error)
	mu      sync.RWMutex
}

// SnippetNodeSpec holds the specifications for creating a SnippetNode.
type SnippetNodeSpec struct {
	scheme.SpecMeta `map:",inline"`
	Lang            string `map:"lang,omitempty"`
	Code            string `map:"code"`
}

const KindSnippet = "snippet"

// NewSnippetNode creates a new SnippetNode with the specified language.Language and code.
func NewSnippetNode(code, lang string) (*SnippetNode, error) {
	if lang == "" {
		lang = language.Detect(code)
	}

	transform, err := language.CompileTransformWithPrimitive(code, lang)
	if err != nil {
		return nil, err
	}

	n := &SnippetNode{
		program: transform,
	}
	n.OneToOneNode = node.NewOneToOneNode(n.action)

	return n, nil
}

func (n *SnippetNode) action(_ *process.Process, inPck *packet.Packet) (*packet.Packet, *packet.Packet) {
	n.mu.RLock()
	defer n.mu.RUnlock()

	if outPayload, err := n.program(inPck.Payload()); err != nil {
		return nil, packet.WithError(err)
	} else {
		return packet.New(outPayload), nil
	}
}

// NewSnippetNodeCodec creates a new codec for SnippetNodeSpec.
func NewSnippetNodeCodec() scheme.Codec {
	return scheme.CodecWithType(func(spec *SnippetNodeSpec) (node.Node, error) {
		return NewSnippetNode(spec.Code, spec.Lang)
	})
}
