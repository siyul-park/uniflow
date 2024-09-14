package control

import (
	"reflect"

	"github.com/siyul-park/uniflow/ext/pkg/language"
	"github.com/siyul-park/uniflow/pkg/node"
	"github.com/siyul-park/uniflow/pkg/packet"
	"github.com/siyul-park/uniflow/pkg/process"
	"github.com/siyul-park/uniflow/pkg/scheme"
	"github.com/siyul-park/uniflow/pkg/spec"
	"github.com/siyul-park/uniflow/pkg/types"
)

// IfNodeSpec defines the specifications for creating an IfNode.
type IfNodeSpec struct {
	spec.Meta `map:",inline"`
	When      string `map:"when"`
}

// IfNode evaluates a condition and routes packets based on the result.
type IfNode struct {
	*node.OneToManyNode
	condition func(any) (bool, error)
}

const KindIf = "if"

// NewIfNodeCodec creates a new codec for IfNodeSpec.
func NewIfNodeCodec(compiler language.Compiler) scheme.Codec {
	return scheme.CodecWithType(func(spec *IfNodeSpec) (node.Node, error) {
		program, err := compiler.Compile(spec.When)
		if err != nil {
			return nil, err
		}
		return NewIfNode(func(env any) (bool, error) {
			res, err := program.Run(env)
			if err != nil {
				return false, err
			}
			return !reflect.ValueOf(res).IsZero(), nil
		}), nil
	})
}

// NewIfNode creates a new IfNode instance.
func NewIfNode(condition func(any) (bool, error)) *IfNode {
	n := &IfNode{condition: condition}
	n.OneToManyNode = node.NewOneToManyNode(n.action)
	return n
}

func (n *IfNode) action(_ *process.Process, inPck *packet.Packet) ([]*packet.Packet, *packet.Packet) {
	inPayload := inPck.Payload()
	input := types.InterfaceOf(inPayload)

	ok, err := n.condition(input)
	if err != nil {
		return nil, packet.New(types.NewError(err))
	}
	if ok {
		return []*packet.Packet{inPck, nil}, nil
	}
	return []*packet.Packet{nil, inPck}, nil
}
