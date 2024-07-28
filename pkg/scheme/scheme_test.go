package scheme

import (
	"reflect"
	"testing"

	"github.com/go-faker/faker/v4"
	"github.com/gofrs/uuid"
	"github.com/siyul-park/uniflow/pkg/node"
	"github.com/siyul-park/uniflow/pkg/secret"
	"github.com/siyul-park/uniflow/pkg/spec"
	"github.com/stretchr/testify/assert"
)

func TestScheme_KnownType(t *testing.T) {
	s := New()
	kind := faker.UUIDHyphenated()

	s.AddKnownType(kind, &spec.Meta{})

	typ, ok := s.KnownType(kind)
	assert.True(t, ok)
	assert.Equal(t, reflect.TypeOf(&spec.Meta{}), typ)
}

func TestScheme_Codec(t *testing.T) {
	s := New()
	kind := faker.UUIDHyphenated()

	c := CodecFunc(func(spec spec.Spec) (node.Node, error) {
		return node.NewOneToOneNode(nil), nil
	})

	s.AddCodec(kind, c)

	_, ok := s.Codec(kind)
	assert.True(t, ok)
}

func TestScheme_Bind(t *testing.T) {
	s := New()
	kind := faker.UUIDHyphenated()

	s.AddKnownType(kind, &spec.Meta{})

	secret := &secret.Secret{
		ID:   uuid.Must(uuid.NewV7()),
		Data: "foo",
	}

	meta := &spec.Meta{
		ID:   uuid.Must(uuid.NewV7()),
		Kind: kind,
		Env: map[string]spec.Secret{
			"FOO": {
				ID:    secret.ID,
				Value: "{{ . }}",
			},
		},
	}

	bind, err := s.Bind(meta, secret)
	assert.NoError(t, err)
	assert.Equal(t, bind.GetEnv()["FOO"].Value, "foo")
	assert.True(t, s.IsBound(bind, secret))
}

func TestScheme_Decode(t *testing.T) {
	s := New()
	kind := faker.UUIDHyphenated()

	s.AddKnownType(kind, &spec.Meta{})

	meta := &spec.Unstructured{
		Meta: spec.Meta{
			ID:   uuid.Must(uuid.NewV7()),
			Kind: kind,
			Env: map[string]spec.Secret{
				"FOO": {
					Value: "foo",
				},
			},
		},
		Fields: map[string]any{
			"foo": "{{ .FOO }}",
		},
	}

	structured, err := s.Decode(meta)
	assert.NoError(t, err)
	assert.Equal(t, structured.GetID(), meta.GetID())
	assert.IsType(t, structured, &spec.Meta{})
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
