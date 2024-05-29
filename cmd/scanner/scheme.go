package scanner

import (
	"github.com/siyul-park/uniflow/pkg/object"
	"github.com/siyul-park/uniflow/pkg/scheme"
)

// SpecCodecOptions holds options for creating a SpecCodec.
type SpecCodecOptions struct {
	Scheme    *scheme.Scheme
	Namespace string
}

// SpecCodec is responsible for decoding raw data into scheme.Spec instances.
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

// Decode decodes raw data into a scheme.Spec instance.
func (c *SpecCodec) Decode(data any) (scheme.Spec, error) {
	doc, err := object.MarshalBinary(data)
	if err != nil {
		return nil, err
	}

	unstructured := scheme.NewUnstructured(doc.(object.Map))

	if unstructured.GetNamespace() == "" {
		if c.namespace != "" {
			unstructured.SetNamespace(c.namespace)
		} else {
			unstructured.SetNamespace(scheme.DefaultNamespace)
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
