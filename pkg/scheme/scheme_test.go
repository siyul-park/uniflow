package scheme

import (
	"testing"

	"github.com/go-faker/faker/v4"
	"github.com/gofrs/uuid"
	"github.com/siyul-park/uniflow/pkg/node"
	"github.com/siyul-park/uniflow/pkg/spec"
	"github.com/stretchr/testify/assert"
)

func TestScheme_KnownType(t *testing.T) {
	s := New()
	kind := faker.UUIDHyphenated()

	ok := s.AddKnownType(kind, &spec.Meta{})
	assert.True(t, ok)
	assert.NotNil(t, s.KnownType(kind))

	ok = s.AddKnownType(kind, &spec.Meta{})
	assert.False(t, ok)

	ok = s.RemoveKnownType(kind)
	assert.True(t, ok)
	assert.Nil(t, s.KnownType(kind))

	ok = s.RemoveKnownType(kind)
	assert.False(t, ok)
}

func TestScheme_Codec(t *testing.T) {
	s := New()
	kind := faker.UUIDHyphenated()

	c := CodecFunc(func(spec spec.Spec) (node.Node, error) {
		return node.NewOneToOneNode(nil), nil
	})

	ok := s.AddCodec(kind, c)
	assert.True(t, ok)
	assert.NotNil(t, s.Codec(kind))

	ok = s.AddCodec(kind, c)
	assert.False(t, ok)

	ok = s.RemoveCodec(kind)
	assert.True(t, ok)
	assert.Nil(t, s.Codec(kind))

	ok = s.RemoveCodec(kind)
	assert.False(t, ok)
}

func TestScheme_Decode(t *testing.T) {
	s := New()
	kind := faker.UUIDHyphenated()

	s.AddKnownType(kind, &spec.Meta{})

	meta := &spec.Unstructured{
		Meta: spec.Meta{
			ID:   uuid.Must(uuid.NewV7()),
			Kind: kind,
			Env: map[string][]spec.Value{
				"FOO": {
					{
						Value: "foo",
					},
				},
			},
		},
		Fields: map[string]any{
			"foo": "{{ .FOO }}",
		},
	}

	structured, err := s.Decode(meta)
	assert.NoError(t, err)
	assert.Equal(t, meta.GetID(), structured.GetID())
	assert.IsType(t, &spec.Meta{}, structured)
}

func TestScheme_Compile(t *testing.T) {
	s := New()
	kind := faker.UUIDHyphenated()

	s.AddKnownType(kind, &spec.Meta{})
	s.AddCodec(kind, CodecFunc(func(spec spec.Spec) (node.Node, error) {
		return node.NewOneToOneNode(nil), nil
	}))

	n, err := s.Compile(&spec.Meta{
		Kind: kind,
	})
	assert.NoError(t, err)
	assert.NotNil(t, n)
}
