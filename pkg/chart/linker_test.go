package chart

import (
	"testing"

	"github.com/go-faker/faker/v4"
	"github.com/gofrs/uuid"
	"github.com/siyul-park/uniflow/pkg/node"
	"github.com/siyul-park/uniflow/pkg/resource"
	"github.com/siyul-park/uniflow/pkg/scheme"
	"github.com/siyul-park/uniflow/pkg/spec"
	"github.com/siyul-park/uniflow/pkg/value"
	"github.com/stretchr/testify/assert"
)

func TestLinker_Link(t *testing.T) {
	s := scheme.New()
	kind := faker.UUIDHyphenated()

	s.AddKnownType(kind, &spec.Meta{})
	s.AddCodec(kind, scheme.CodecFunc(func(spec spec.Spec) (node.Node, error) {
		return node.NewOneToOneNode(nil), nil
	}))

	l := NewLinker(s)

	scrt := &value.Value{ID: uuid.Must(uuid.NewV7())}
	chrt := &Chart{
		ID:        uuid.Must(uuid.NewV7()),
		Namespace: resource.DefaultNamespace,
		Name:      faker.UUIDHyphenated(),
		Specs: []*spec.Unstructured{
			{
				Meta: spec.Meta{
					Kind: kind,
					Name: "dummy",
				},
			},
		},
		Env: map[string][]spec.Value{
			"key1": {
				{
					ID:   scrt.GetID(),
					Data: faker.UUIDHyphenated(),
				},
			},
			"key2": {
				{
					Data: "{{ .id }}",
				},
			},
		},
		Inbounds: map[string][]spec.Port{
			node.PortIn: {
				{
					Name: "dummy",
					Port: node.PortIn,
				},
			},
		},
	}

	meta := &spec.Meta{
		Kind:      chrt.GetName(),
		Namespace: resource.DefaultNamespace,
	}

	err := l.Link(chrt)
	assert.NoError(t, err)
	assert.Contains(t, s.Kinds(), chrt.GetName())

	n, err := s.Compile(meta)
	assert.NoError(t, err)
	assert.NotNil(t, n)
}

func TestLinker_Unlink(t *testing.T) {
	s := scheme.New()

	l := NewLinker(s)

	chrt := &Chart{
		ID:        uuid.Must(uuid.NewV7()),
		Namespace: resource.DefaultNamespace,
		Name:      faker.UUIDHyphenated(),
	}

	l.Link(chrt)

	err := l.Unlink(chrt)
	assert.NoError(t, err)
	assert.NotContains(t, s.Kinds(), chrt.GetName())
}
