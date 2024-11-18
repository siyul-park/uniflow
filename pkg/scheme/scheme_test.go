package scheme

import (
	"github.com/siyul-park/uniflow/pkg/secret"
	"testing"

	"github.com/go-faker/faker/v4"
	"github.com/gofrs/uuid"
	"github.com/siyul-park/uniflow/pkg/node"
	"github.com/siyul-park/uniflow/pkg/spec"
	"github.com/stretchr/testify/assert"
)

func TestScheme_Kinds(t *testing.T) {
	s := New()
	kind := faker.UUIDHyphenated()

	c := CodecFunc(func(spec spec.Spec) (node.Node, error) {
		return node.NewOneToOneNode(nil), nil
	})

	s.AddKnownType(kind, &spec.Meta{})
	s.AddCodec(kind, c)

	kinds := s.Kinds()
	assert.Contains(t, kinds, kind)
}

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

func TestScheme_IsBound(t *testing.T) {
	s := New()

	sec1 := &secret.Secret{
		ID: uuid.Must(uuid.NewV7()),
	}
	sec2 := &secret.Secret{
		ID: uuid.Must(uuid.NewV7()),
	}

	meta := &spec.Meta{
		ID:   uuid.Must(uuid.NewV7()),
		Kind: faker.UUIDHyphenated(),
		Env: map[string][]spec.Value{
			"FOO": {
				{
					ID:   sec1.ID,
					Data: "foo",
				},
			},
		},
	}

	assert.True(t, s.IsBound(meta, sec1))
	assert.False(t, s.IsBound(meta, sec2))
}

func TestScheme_Bind(t *testing.T) {
	s := New()
	kind := faker.UUIDHyphenated()

	s.AddKnownType(kind, &spec.Meta{})

	sec := &secret.Secret{
		ID:   uuid.Must(uuid.NewV7()),
		Data: faker.UUIDHyphenated(),
	}
	meta := &spec.Unstructured{
		Meta: spec.Meta{
			ID:   uuid.Must(uuid.NewV7()),
			Kind: kind,
			Env: map[string][]spec.Value{
				"FOO": {
					{
						ID:   sec.ID,
						Data: "{{ . }}",
					},
				},
			},
		},
		Fields: map[string]any{
			"foo": "{{ .FOO }}",
		},
	}

	bind, err := s.Bind(meta, sec)
	assert.NoError(t, err)
	assert.Equal(t, sec.Data, bind.GetEnv()["FOO"][0].Data)
}

func TestScheme_Decode(t *testing.T) {
	s := New()
	kind := faker.UUIDHyphenated()

	s.AddKnownType(kind, &spec.Meta{})

	meta := &spec.Unstructured{
		Meta: spec.Meta{
			ID:   uuid.Must(uuid.NewV7()),
			Kind: kind,
		},
	}

	decode, err := s.Decode(meta)
	assert.NoError(t, err)
	assert.IsType(t, decode, &spec.Meta{})
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
