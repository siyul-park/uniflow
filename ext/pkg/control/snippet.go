package control

import (
	"context"
	"time"

	"github.com/siyul-park/uniflow/ext/pkg/language"
	"github.com/siyul-park/uniflow/pkg/node"
	"github.com/siyul-park/uniflow/pkg/packet"
	"github.com/siyul-park/uniflow/pkg/process"
	"github.com/siyul-park/uniflow/pkg/scheme"
	"github.com/siyul-park/uniflow/pkg/spec"
	"github.com/siyul-park/uniflow/pkg/types"
)

// SnippetNodeSpec defines the specifications for creating a SnippetNode.
type SnippetNodeSpec struct {
	spec.Meta `map:",inline"`
	Language  string        `map:"language" validate:"required"`
	Code      string        `map:"code" validate:"required"`
	Timeout   time.Duration `map:"timeout,omitempty"`
}

// SnippetNode represents a node that executes code snippets in various languages.
type SnippetNode struct {
	*node.OneToOneNode
	fn func(context.Context, any) (any, error)
}

const KindSnippet = "snippet"

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

		return NewSnippetNode(language.Function[any, any](language.Timeout(program, spec.Timeout))), nil
	})
}

// NewSnippetNode creates a new SnippetNode with the specified language.Language and code.
func NewSnippetNode(fn func(context.Context, any) (any, error)) *SnippetNode {
	n := &SnippetNode{fn: fn}
	n.OneToOneNode = node.NewOneToOneNode(n.action)
	return n
}

func (n *SnippetNode) action(proc *process.Process, inPck *packet.Packet) (*packet.Packet, *packet.Packet) {
	inPayload := inPck.Payload()
	input := types.InterfaceOf(inPayload)

	if output, err := n.fn(proc.Context(), input); err != nil {
		return nil, packet.New(types.NewError(err))
	} else if outPayload, err := types.Marshal(output); err != nil {
		return nil, packet.New(types.NewError(err))
	} else {
		return packet.New(outPayload), nil
	}
}
