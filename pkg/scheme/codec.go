package scheme

import (
	"github.com/pkg/errors"
	"github.com/siyul-park/uniflow/pkg/encoding"
	"github.com/siyul-park/uniflow/pkg/node"
	"github.com/siyul-park/uniflow/pkg/spec"
)

// Codec defines the interface for converting a spec.Spec into a node.Node.
type Codec interface {
	// Compile converts the given spec.Spec into a node.Node.
	Compile(sp spec.Spec) (node.Node, error)
}

type codec struct {
	compile func(sp spec.Spec) (node.Node, error)
}

var _ Codec = (*codec)(nil)

// CodecFunc takes a compile function and returns a struct that implements the Codec interface.
func CodecFunc(compile func(sp spec.Spec) (node.Node, error)) Codec {
	return &codec{compile: compile}
}

// CodecWithType creates a Codec that works with a specific type T.
func CodecWithType[T spec.Spec](compile func(spec T) (node.Node, error)) Codec {
	return CodecFunc(func(spec spec.Spec) (node.Node, error) {
		if converted, ok := spec.(T); ok {
			return compile(converted)
		}
		return nil, errors.WithStack(encoding.ErrUnsupportedType)
	})
}

func (c *codec) Compile(sp spec.Spec) (node.Node, error) {
	return c.compile(sp)
}
