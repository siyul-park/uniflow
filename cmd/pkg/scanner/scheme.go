package scanner

import (
	"github.com/siyul-park/uniflow/pkg/scheme"
	"github.com/siyul-park/uniflow/pkg/spec"
	"github.com/siyul-park/uniflow/pkg/types"
)

// SpecCodecOptions holds options for creating a SpecCodec.
type SpecCodecOptions struct {
	Scheme    *scheme.Scheme
	Namespace string
}

// SpecCodec is responsible for decoding raw data into spec.Spec instances.
type SpecCodec struct {
	scheme    *scheme.Scheme
	namespace string
}

// NewSpecCodec creates a new SpecCodec instance with the provided options.
func NewSpecCodec(opts ...SpecCodecOptions) *SpecCodec {
	var scheme *scheme.Scheme
	var namespace string

	for _, opt := range opts {
		if opt.Scheme != nil {
			scheme = opt.Scheme
		}
		if opt.Namespace != "" {
			namespace = opt.Namespace
		}
	}

	return &SpecCodec{
		scheme:    scheme,
		namespace: namespace,
	}
}

// Decode decodes raw data into a spec.Spec instance.
func (c *SpecCodec) Decode(data any) (spec.Spec, error) {
	doc, err := types.BinaryEncoder.Encode(data)
	if err != nil {
		return nil, err
	}

	unstructured := spec.NewUnstructured(doc.(types.Map))

	if unstructured.GetNamespace() == "" {
		if c.namespace != "" {
			unstructured.SetNamespace(c.namespace)
		} else {
			unstructured.SetNamespace(spec.DefaultNamespace)
		}
	}

	if c.scheme == nil {
		return unstructured, nil
	}

	return c.scheme.Decode(unstructured)
}
