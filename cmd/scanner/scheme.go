package scanner

import (
	"github.com/siyul-park/uniflow/pkg/object"
	"github.com/siyul-park/uniflow/pkg/spec"
)

// SpecCodecOptions holds options for creating a SpecCodec.
type SpecCodecOptions struct {
	Scheme    *spec.Scheme
	Namespace string
}

// SpecCodec is responsible for decoding raw data into spec.Spec instances.
type SpecCodec struct {
	scheme    *spec.Scheme
	namespace string
}

// NewSpecCodec creates a new SpecCodec instance with the provided options.
func NewSpecCodec(opts ...SpecCodecOptions) *SpecCodec {
	var scheme *spec.Scheme
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
	doc, err := object.MarshalBinary(data)
	if err != nil {
		return nil, err
	}

	unstructured := spec.NewUnstructured(doc.(*object.Map))

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

	if spec, ok := c.scheme.Spec(unstructured.GetKind()); !ok {
		return unstructured, nil
	} else if err := object.Unmarshal(doc, spec); err != nil {
		return nil, err
	} else {
		return spec, nil
	}
}
