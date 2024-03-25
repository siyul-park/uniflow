package scheme

import (
	"github.com/gofrs/uuid"
	"reflect"
	"testing"

	"github.com/go-faker/faker/v4"
	"github.com/siyul-park/uniflow/pkg/node"
	"github.com/stretchr/testify/assert"
)

func TestScheme_KnownType(t *testing.T) {
	s := New()
	kind := faker.UUIDHyphenated()

	s.AddKnownType(kind, &SpecMeta{})

	typ, ok := s.KnownType(kind)
	assert.True(t, ok)
	assert.Equal(t, reflect.TypeOf(&SpecMeta{}), typ)
}

func TestScheme_Codec(t *testing.T) {
	s := New()
	kind := faker.UUIDHyphenated()

	c := CodecFunc(func(spec Spec) (node.Node, error) {
		return node.NewOneToOneNode(nil), nil
	})

	s.AddCodec(kind, c)

	_, ok := s.Codec(kind)
	assert.True(t, ok)
}

func TestScheme_Unstructured(t *testing.T) {
	s := New()
	kind := faker.UUIDHyphenated()

	s.AddKnownType(kind, &SpecMeta{})

	spec := &SpecMeta{
		ID:   uuid.Must(uuid.NewV7()),
		Kind: kind,
	}

	unstructured, err := s.Unstructured(spec)
	assert.NoError(t, err)
	assert.Equal(t, unstructured.GetID(), spec.GetID())
	assert.IsType(t, unstructured, &Unstructured{})
}

func TestScheme_Structured(t *testing.T) {
	s := New()
	kind := faker.UUIDHyphenated()

	s.AddKnownType(kind, &SpecMeta{})

	spec := &SpecMeta{
		ID:   uuid.Must(uuid.NewV7()),
		Kind: kind,
	}

	structured, err := s.Structured(spec)
	assert.NoError(t, err)
	assert.Equal(t, structured.GetID(), spec.GetID())
	assert.IsType(t, structured, &SpecMeta{})
}

func TestScheme_Spec(t *testing.T) {
	s := New()
	kind := faker.UUIDHyphenated()

	s.AddKnownType(kind, &SpecMeta{})

	spec, ok := s.Spec(kind)
	assert.True(t, ok)
	assert.IsType(t, spec, &SpecMeta{})
}

func TestScheme_Decode(t *testing.T) {
	s := New()
	kind := faker.UUIDHyphenated()

	s.AddKnownType(kind, &SpecMeta{})
	s.AddCodec(kind, CodecFunc(func(spec Spec) (node.Node, error) {
		return node.NewOneToOneNode(nil), nil
	}))

	n, err := s.Decode(&SpecMeta{})
	assert.NoError(t, err)
	assert.NotNil(t, n)
}

func TestScheme_Kinds(t *testing.T) {
	s := New()
	kind := faker.UUIDHyphenated()

	s.AddKnownType(kind, &SpecMeta{})

	kinds := s.Kinds(&SpecMeta{})
	assert.Len(t, kinds, 1)
	assert.Equal(t, kind, kinds[0])
}
