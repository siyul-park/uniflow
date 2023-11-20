package scheme

import (
	"github.com/pkg/errors"
	"github.com/siyul-park/uniflow/internal/encoding"
	"github.com/siyul-park/uniflow/pkg/node"
)

type (
	// Codec is the interface for decoding Spec to node.Node.
	Codec interface {
		Decode(spec Spec) (node.Node, error)
	}

	CodecFunc func(spec Spec) (node.Node, error)
)

func CodecWithType[T Spec](decode func(spec T) (node.Node, error)) Codec {
	return CodecFunc(func(spec Spec) (node.Node, error) {
		if spec, ok := spec.(T); !ok {
			return nil, errors.WithStack(encoding.ErrUnsupportedValue)
		} else {
			return decode(spec)
		}
	})
}

func (c CodecFunc) Decode(spec Spec) (node.Node, error) {
	return c(spec)
}
