package scheme

import (
	"github.com/pkg/errors"
	"github.com/siyul-park/uniflow/pkg/encoding"
	"github.com/siyul-park/uniflow/pkg/node"
)

// Codec defines the interface for decoding Spec to node.Node.
type Codec interface {
	Decode(spec Spec) (node.Node, error)
}

// CodecFunc represents a function type that implements the Codec interface.
type CodecFunc func(spec Spec) (node.Node, error)

// CodecWithType creates a new CodecFunc for the specified type T.
func CodecWithType[T Spec](decode func(spec T) (node.Node, error)) Codec {
	return CodecFunc(func(spec Spec) (node.Node, error) {
		if converted, ok := spec.(T); ok {
			return decode(converted)
		}
		return nil, errors.WithStack(encoding.ErrInvalidValue)
	})
}

// Decode implements the Decode method for CodecFunc.
func (c CodecFunc) Decode(spec Spec) (node.Node, error) {
	return c(spec)
}
