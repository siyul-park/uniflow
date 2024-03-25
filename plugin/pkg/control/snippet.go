package control

import (
	"sync"

	"github.com/siyul-park/uniflow/pkg/node"
	"github.com/siyul-park/uniflow/pkg/packet"
	"github.com/siyul-park/uniflow/pkg/process"
	"github.com/siyul-park/uniflow/pkg/scheme"
	"github.com/siyul-park/uniflow/plugin/internal/language"
)

// SnippetNode represents a node that executes code snippets in various language.
type SnippetNode struct {
	*node.OneToOneNode
	mu sync.RWMutex
}

// SnippetNodeSpec holds the specifications for creating a SnippetNode.
type SnippetNodeSpec struct {
	scheme.SpecMeta `map:",inline"`
	Lang            string `map:"lang,omitempty"`
	Code            string `map:"code"`
}

const KindSnippet = "snippet"

// NewSnippetNode creates a new SnippetNode with the specified language.Language and code.
func NewSnippetNode(lang, code string) (*SnippetNode, error) {
	if lang == "" {
		lang = language.Detect(code)
	}

	n := &SnippetNode{}
	if action, err := n.compile(code, lang); err != nil {
		return nil, err
	} else {
		n.OneToOneNode = node.NewOneToOneNode(action)
	}
	return n, nil
}

func (n *SnippetNode) compile(code, lang string) (func(*process.Process, *packet.Packet) (*packet.Packet, *packet.Packet), error) {
	n.mu.RLock()
	defer n.mu.RUnlock()

	transform, err := language.CompileTransformWithPrimitive(code, lang)
	if err != nil {
		return nil, err
	}

	return func(proc *process.Process, inPck *packet.Packet) (*packet.Packet, *packet.Packet) {
		if outPayload, err := transform(inPck.Payload()); err != nil {
			return nil, packet.WithError(err, inPck)
		} else {
			return packet.New(outPayload), nil
		}
	}, nil
}

// NewSnippetNodeCodec creates a new codec for SnippetNodeSpec.
func NewSnippetNodeCodec() scheme.Codec {
	return scheme.CodecWithType[*SnippetNodeSpec](func(spec *SnippetNodeSpec) (node.Node, error) {
		return NewSnippetNode(spec.Lang, spec.Code)
	})
}
