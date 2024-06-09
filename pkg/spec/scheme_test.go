package spec

import (
	"reflect"
	"testing"

	"github.com/gofrs/uuid"

	"github.com/go-faker/faker/v4"
	"github.com/siyul-park/uniflow/pkg/node"
	"github.com/stretchr/testify/assert"
)

func TestScheme_KnownType(t *testing.T) {
	s := NewScheme()
	kind := faker.UUIDHyphenated()

	s.AddKnownType(kind, &Meta{})

	typ, ok := s.KnownType(kind)
	assert.True(t, ok)
	assert.Equal(t, reflect.TypeOf(&Meta{}), typ)
}

func TestScheme_Codec(t *testing.T) {
	s := NewScheme()
	kind := faker.UUIDHyphenated()

	c := CodecFunc(func(spec Spec) (node.Node, error) {
		return node.NewOneToOneNode(nil), nil
	})

	s.AddCodec(kind, c)

	_, ok := s.Codec(kind)
	assert.True(t, ok)
}

func TestScheme_Unstructured(t *testing.T) {
	s := NewScheme()
	kind := faker.UUIDHyphenated()

	s.AddKnownType(kind, &Meta{})

	spec := &Meta{
		ID:   uuid.Must(uuid.NewV7()),
		Kind: kind,
	}

	unstructured, err := s.Unstructured(spec)
	assert.NoError(t, err)
	assert.Equal(t, unstructured.GetID(), spec.GetID())
	assert.IsType(t, unstructured, &Unstructured{})
}

func TestScheme_Structured(t *testing.T) {
	s := NewScheme()
	kind := faker.UUIDHyphenated()

	s.AddKnownType(kind, &Meta{})

	spec := &Meta{
		ID:   uuid.Must(uuid.NewV7()),
		Kind: kind,
	}

	structured, err := s.Structured(spec)
	assert.NoError(t, err)
	assert.Equal(t, structured.GetID(), spec.GetID())
	assert.IsType(t, structured, &Meta{})
}

func TestScheme_Spec(t *testing.T) {
	s := NewScheme()
	kind := faker.UUIDHyphenated()

	s.AddKnownType(kind, &Meta{})

	spec, ok := s.Spec(kind)
	assert.True(t, ok)
	assert.IsType(t, spec, &Meta{})
}

func TestScheme_Decode(t *testing.T) {
	s := NewScheme()
	kind := faker.UUIDHyphenated()

	s.AddKnownType(kind, &Meta{})
	s.AddCodec(kind, CodecFunc(func(spec Spec) (node.Node, error) {
		return node.NewOneToOneNode(nil), nil
	}))

	n, err := s.Decode(&Meta{})
	assert.NoError(t, err)
	assert.NotNil(t, n)
}

func TestScheme_Kinds(t *testing.T) {
	s := NewScheme()
	kind := faker.UUIDHyphenated()

	s.AddKnownType(kind, &Meta{})

	kinds := s.Kinds(&Meta{})
	assert.Len(t, kinds, 1)
	assert.Equal(t, kind, kinds[0])
}
