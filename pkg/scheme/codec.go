package scheme

import (
	"github.com/pkg/errors"
	"github.com/siyul-park/uniflow/pkg/encoding"
	"github.com/siyul-park/uniflow/pkg/node"
	"github.com/siyul-park/uniflow/pkg/spec"
)

// Codec defines the interface for decoding Spec to node.Node.
type Codec interface {
	Compile(spc spec.Spec) (node.Node, error)
}

// CodecFunc represents a function type that implements the Codec interface.
type CodecFunc func(spc spec.Spec) (node.Node, error)

// CodecWithType creates a new CodecFunc for the specified type T.
func CodecWithType[T spec.Spec](compile func(spec T) (node.Node, error)) Codec {
	return CodecFunc(func(spec spec.Spec) (node.Node, error) {
		if converted, ok := spec.(T); ok {
			return compile(converted)
		}
		return nil, errors.WithStack(encoding.ErrInvalidArgument)
	})
}

// Compile implements the Compile method for CodecFunc.
func (f CodecFunc) Compile(spc spec.Spec) (node.Node, error) {
	return f(spc)
}
