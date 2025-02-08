package control

import (
	"context"

	"github.com/pkg/errors"
	"github.com/siyul-park/uniflow/ext/pkg/language"
	"github.com/siyul-park/uniflow/pkg/encoding"
	"github.com/siyul-park/uniflow/pkg/node"
	"github.com/siyul-park/uniflow/pkg/packet"
	"github.com/siyul-park/uniflow/pkg/process"
	"github.com/siyul-park/uniflow/pkg/scheme"
	"github.com/siyul-park/uniflow/pkg/spec"
	"github.com/siyul-park/uniflow/pkg/types"
)

const (
	// KindAssert is the kind value for assert nodes.
	KindAssert = "assert"
)

// AssertNodeSpec represents an assert node specification.
type AssertNodeSpec struct {
	spec.Meta `map:",inline"`
	Condition string `map:"condition"` // The condition to assert
}

// AssertNode represents a node that asserts a condition.
type AssertNode struct {
	*node.OneToOneNode
	compiler language.Compiler
	when     language.Program
}

// NewAssertNodeCodec creates a new codec for AssertNodeSpec.
func NewAssertNodeCodec(compiler language.Compiler) scheme.Codec {
	return scheme.CodecWithType(func(spec *AssertNodeSpec) (node.Node, error) {
		if spec.Condition == "" {
			return nil, errors.WithStack(encoding.ErrUnsupportedValue)
		}

		when, err := compiler.Compile(spec.Condition)
		if err != nil {
			return nil, err
		}

		return NewAssertNode(compiler, when), nil
	})
}

// NewAssertNode creates a new assert node.
func NewAssertNode(compiler language.Compiler, when language.Program) node.Node {
	n := &AssertNode{
		compiler: compiler,
		when:     when,
	}

	n.OneToOneNode = node.NewOneToOneNode(n.process)
	return n
}

func (n *AssertNode) process(_ *process.Process, inPck *packet.Packet) (*packet.Packet, *packet.Packet) {
	if inPck == nil {
		return nil, nil
	}

	result, err := n.when.Run(context.TODO(), inPck.Payload())
	if err != nil {
		return nil, packet.New(types.NewError(err))
	}

	ok, isOk := result.(bool)
	if !isOk {
		return nil, packet.New(types.NewError(errors.New("assertion result must be a boolean")))
	}

	if !ok {
		return nil, packet.New(types.NewError(errors.New("assertion failed")))
	}

	return inPck, nil
}

// AddAssertToScheme returns a function that adds node types and codecs to the provided spec.
func AddAssertToScheme(compiler language.Compiler) scheme.Register {
	return scheme.RegisterFunc(func(s *scheme.Scheme) error {
		s.AddKnownType(KindAssert, &AssertNodeSpec{})
		s.AddCodec(KindAssert, NewAssertNodeCodec(compiler))
		return nil
	})
}
