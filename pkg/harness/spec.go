package harness

import (
	"fmt"
	"strings"

	"github.com/siyul-park/uniflow/pkg/node"
	"github.com/siyul-park/uniflow/pkg/packet"
	"github.com/siyul-park/uniflow/pkg/scheme"
	"github.com/siyul-park/uniflow/pkg/spec"
	"github.com/siyul-park/uniflow/pkg/types"
)

const (
	KindSourceNode    = "source"
	KindTransformNode = "transform"
	KindAssertNode    = "assert"
)

// SourceNodeSpec defines the configuration for a source node
type SourceNodeSpec struct {
	spec.Meta `map:",inline"`
	Data      interface{} `map:"data"`
}

// TransformNodeSpec defines the configuration for a transform node
type TransformNodeSpec struct {
	spec.Meta `map:",inline"`
	Transform string `map:"transform"` // JavaScript function as string
}

// AssertNodeSpec defines the configuration for an assert node
type AssertNodeSpec struct {
	spec.Meta     `map:",inline"`
	ExpectedValue interface{}           `map:"expected_value"`
	ResultChan    chan<- *packet.Packet `map:"-"` // Channel for test results
}

// NewSourceNodeCodec creates a codec for converting SourceNodeSpec to SourceNode
func NewSourceNodeCodec() scheme.Codec {
	return scheme.CodecWithType(func(spec *SourceNodeSpec) (node.Node, error) {
		// Convert interface{} to appropriate types.Value based on the actual type
		var value types.Value
		switch v := spec.Data.(type) {
		case string:
			value = types.NewString(v)
		case bool:
			value = types.NewBoolean(v)
		default:
			// For other types, convert to string representation
			value = types.NewString(fmt.Sprintf("%v", v))
		}
		return NewSourceNode(value), nil
	})
}

// NewTransformNodeCodec creates a codec for converting TransformNodeSpec to TransformNode
func NewTransformNodeCodec() scheme.Codec {
	return scheme.CodecWithType(func(spec *TransformNodeSpec) (node.Node, error) {
		// Note: In a real implementation, you would compile the JavaScript function here
		return NewTransformNode(func(v types.Value) (types.Value, error) {
			if str, ok := v.(types.String); ok {
				return types.NewString(strings.ToUpper(str.String())), nil
			}
			return v, nil
		}), nil
	})
}

// NewAssertNodeCodec creates a codec for converting AssertNodeSpec to AssertNode
func NewAssertNodeCodec() scheme.Codec {
	return scheme.CodecWithType(func(spec *AssertNodeSpec) (node.Node, error) {
		// Convert interface{} to appropriate types.Value based on the actual type
		var expected types.Value
		switch v := spec.ExpectedValue.(type) {
		case string:
			expected = types.NewString(v)
		case bool:
			expected = types.NewBoolean(v)
		default:
			// For other types, convert to string representation
			expected = types.NewString(fmt.Sprintf("%v", v))
		}
		return NewAssertNode(spec.ResultChan, func(v types.Value) error {
			if !v.Equal(expected) {
				return types.NewError(nil)
			}
			return nil
		}), nil
	})
}

// AddToScheme registers all harness node types with the scheme
func AddToScheme() scheme.Register {
	return scheme.RegisterFunc(func(s *scheme.Scheme) error {
		// Register source node
		s.AddKnownType(KindSourceNode, &SourceNodeSpec{})
		s.AddCodec(KindSourceNode, NewSourceNodeCodec())

		// Register transform node
		s.AddKnownType(KindTransformNode, &TransformNodeSpec{})
		s.AddCodec(KindTransformNode, NewTransformNodeCodec())

		// Register assert node
		s.AddKnownType(KindAssertNode, &AssertNodeSpec{})
		s.AddCodec(KindAssertNode, NewAssertNodeCodec())

		return nil
	})
}
