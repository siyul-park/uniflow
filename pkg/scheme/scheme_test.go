package scheme

import (
	"testing"

	"github.com/go-faker/faker/v4"
	"github.com/gofrs/uuid"
	"github.com/siyul-park/uniflow/pkg/node"
	"github.com/siyul-park/uniflow/pkg/spec"
	"github.com/stretchr/testify/require"
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
	require.Contains(t, kinds, kind)
}

func TestScheme_KnownType(t *testing.T) {
	s := New()
	kind := faker.UUIDHyphenated()

	ok := s.AddKnownType(kind, &spec.Meta{})
	require.True(t, ok)
	require.NotNil(t, s.KnownType(kind))

	ok = s.AddKnownType(kind, &spec.Meta{})
	require.False(t, ok)

	ok = s.RemoveKnownType(kind)
	require.True(t, ok)
	require.Nil(t, s.KnownType(kind))

	ok = s.RemoveKnownType(kind)
	require.False(t, ok)
}

func TestScheme_Codec(t *testing.T) {
	s := New()
	kind := faker.UUIDHyphenated()

	c := CodecFunc(func(spec spec.Spec) (node.Node, error) {
		return node.NewOneToOneNode(nil), nil
	})

	ok := s.AddCodec(kind, c)
	require.True(t, ok)
	require.NotNil(t, s.Codec(kind))

	ok = s.AddCodec(kind, c)
	require.False(t, ok)

	ok = s.RemoveCodec(kind)
	require.True(t, ok)
	require.Nil(t, s.Codec(kind))

	ok = s.RemoveCodec(kind)
	require.False(t, ok)
}

func TestScheme_Decode(t *testing.T) {
	s := New()
	kind := faker.UUIDHyphenated()

	s.AddKnownType(kind, &spec.Meta{})

	meta := &spec.Unstructured{
		Meta: spec.Meta{
			ID:        uuid.Must(uuid.NewV7()),
			Kind:      kind,
			Namespace: faker.UUIDHyphenated(),
		},
	}

	decode, err := s.Decode(meta)
	require.NoError(t, err)
	require.IsType(t, decode, &spec.Meta{})
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
	require.NoError(t, err)
	require.NotNil(t, n)
}
