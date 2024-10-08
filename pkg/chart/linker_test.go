package chart

import (
	"testing"

	"github.com/go-faker/faker/v4"
	"github.com/gofrs/uuid"
	"github.com/siyul-park/uniflow/pkg/node"
	"github.com/siyul-park/uniflow/pkg/resource"
	"github.com/siyul-park/uniflow/pkg/scheme"
	"github.com/siyul-park/uniflow/pkg/secret"
	"github.com/siyul-park/uniflow/pkg/spec"
	"github.com/stretchr/testify/assert"
)

func TestLinker_Load(t *testing.T) {
	s := scheme.New()
	kind := faker.UUIDHyphenated()

	s.AddKnownType(kind, &spec.Meta{})
	s.AddCodec(kind, scheme.CodecFunc(func(spec spec.Spec) (node.Node, error) {
		return node.NewOneToOneNode(nil), nil
	}))

	l := NewLinker(LinkerConfig{
		Scheme: s,
	})

	sec := &secret.Secret{ID: uuid.Must(uuid.NewV7())}
	chrt := &Chart{
		ID:        uuid.Must(uuid.NewV7()),
		Namespace: resource.DefaultNamespace,
		Name:      faker.UUIDHyphenated(),
		Specs: []spec.Spec{
			&spec.Meta{
				Kind: kind,
				Name: "dummy",
			},
		},
		Env: map[string][]Value{
			"key1": {
				{
					ID:    sec.GetID(),
					Value: faker.Word(),
				},
			},
			"key2": {
				{
					Value: "{{ .id }}",
				},
			},
		},
		Ports: map[string][]Port{
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

	err := l.Load(chrt)
	assert.NoError(t, err)
	assert.Contains(t, s.Kinds(), chrt.GetName())

	n, err := s.Compile(meta)
	assert.NoError(t, err)
	assert.NotNil(t, n)
}

func TestLinker_Unload(t *testing.T) {
	s := scheme.New()

	l := NewLinker(LinkerConfig{
		Scheme: s,
	})

	chrt := &Chart{
		ID:        uuid.Must(uuid.NewV7()),
		Namespace: resource.DefaultNamespace,
		Name:      faker.UUIDHyphenated(),
		Specs:     []spec.Spec{},
	}

	s.AddKnownType(chrt.GetName(), &spec.Meta{})
	s.AddCodec(chrt.GetName(), scheme.CodecFunc(func(spec spec.Spec) (node.Node, error) {
		return node.NewOneToOneNode(nil), nil
	}))

	err := l.Unload(chrt)
	assert.NoError(t, err)
	assert.NotContains(t, s.Kinds(), chrt.GetName())
}
