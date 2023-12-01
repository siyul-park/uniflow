package resource

import (
	"github.com/pkg/errors"
	"github.com/siyul-park/uniflow/pkg/encoding"
	"github.com/siyul-park/uniflow/pkg/primitive"
	"github.com/siyul-park/uniflow/pkg/scheme"
)

type (
	SpecCodecOptions struct {
		Scheme    *scheme.Scheme
		Namespace string
	}

	SpecCodec struct {
		scheme    *scheme.Scheme
		namespace string
	}
)

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

func (c *SpecCodec) Decode(data any) (scheme.Spec, error) {
	doc, err := primitive.MarshalBinary(data)
	if err != nil {
		return nil, err
	}

	unstructured := scheme.NewUnstructured(doc.(*primitive.Map))

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

	spec, ok := c.scheme.New(unstructured.GetKind())
	if !ok {
		return nil, errors.WithStack(encoding.ErrUnsupportedValue)
	}
	if err := unstructured.Unmarshal(spec); err != nil {
		return nil, err
	}
	return spec, nil
}
