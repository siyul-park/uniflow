package testing

import (
	"fmt"

	"github.com/siyul-park/uniflow/pkg/node"
	"github.com/siyul-park/uniflow/pkg/scheme"
	"github.com/siyul-park/uniflow/pkg/spec"
)

const (
	// KindTest is the kind value for test nodes.
	KindTest = "test"
)

// TestNodeSpec represents a test node specification.
type TestNodeSpec struct {
	spec.Meta `map:",inline"`
	Specs     []spec.Spec            `map:"specs" validate:"required"`
	Ports     map[string][]spec.Port `map:"ports,omitempty"`
}

// NewTestNodeSpec creates a new TestNodeSpec.
func NewTestNodeSpec() *TestNodeSpec {
	return &TestNodeSpec{
		Meta: spec.Meta{
			Kind: KindTest,
		},
		Ports: make(map[string][]spec.Port),
	}
}

// NewTestNodeCodec creates a new codec for TestNodeSpec.
func NewTestNodeCodec(s *scheme.Scheme) scheme.Codec {
	return scheme.CodecFunc(func(sp spec.Spec) (node.Node, error) {
		var root TestNodeSpec
		if err := spec.As(sp, &root); err != nil {
			return nil, fmt.Errorf("failed to convert spec to TestNodeSpec: %v", err)
		}

		// Decode each child spec through scheme to set defaults and validate
		for i, childSpec := range root.Specs {
			decoded, err := s.Decode(childSpec)
			if err != nil {
				return nil, fmt.Errorf("failed to decode child spec at index %d: %v", i, err)
			}
			root.Specs[i] = decoded
		}

		node, err := NewTestNode(&root, s)
		if err != nil {
			return nil, fmt.Errorf("failed to create test node: %v", err)
		}
		return node, nil
	})
}

// AddToScheme returns a function that adds node types and codecs to the provided spec.
func AddToScheme() scheme.Register {
	return scheme.RegisterFunc(func(s *scheme.Scheme) error {
		if !s.AddKnownType(KindTest, &TestNodeSpec{}) {
			return fmt.Errorf("failed to add test node spec")
		}
		if !s.AddCodec(KindTest, NewTestNodeCodec(s)) {
			return fmt.Errorf("failed to add test node codec")
		}
		return nil
	})
}
