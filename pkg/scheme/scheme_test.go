package scheme

import (
	"reflect"
	"testing"

	"github.com/gofrs/uuid"

	"github.com/go-faker/faker/v4"
	"github.com/siyul-park/uniflow/pkg/node"
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

func TestScheme_Unstructured(t *testing.T) {
	s := New()
	kind := faker.UUIDHyphenated()

	s.AddKnownType(kind, &spec.Meta{})

	meta := &spec.Meta{
		ID:   uuid.Must(uuid.NewV7()),
		Kind: kind,
	}

	unstructured, err := s.Unstructured(meta)
	assert.NoError(t, err)
	assert.Equal(t, unstructured.GetID(), meta.GetID())
	assert.IsType(t, unstructured, &spec.Unstructured{})
}

func TestScheme_Structured(t *testing.T) {
	s := New()
	kind := faker.UUIDHyphenated()

	s.AddKnownType(kind, &spec.Meta{})

	meta := &spec.Meta{
		ID:   uuid.Must(uuid.NewV7()),
		Kind: kind,
	}

	structured, err := s.Structured(meta)
	assert.NoError(t, err)
	assert.Equal(t, structured.GetID(), meta.GetID())
	assert.IsType(t, structured, &spec.Meta{})
}

func TestScheme_Spec(t *testing.T) {
	s := New()
	kind := faker.UUIDHyphenated()

	s.AddKnownType(kind, &spec.Meta{})

	meta, ok := s.Spec(kind)
	assert.True(t, ok)
	assert.IsType(t, meta, &spec.Meta{})
}

func TestScheme_Decode(t *testing.T) {
	s := New()
	kind := faker.UUIDHyphenated()

	s.AddKnownType(kind, &spec.Meta{})
	s.AddCodec(kind, CodecFunc(func(spec spec.Spec) (node.Node, error) {
		return node.NewOneToOneNode(nil), nil
	}))

	n, err := s.Decode(&spec.Meta{})
	assert.NoError(t, err)
	assert.NotNil(t, n)
}

func TestScheme_Kinds(t *testing.T) {
	s := New()
	kind := faker.UUIDHyphenated()

	s.AddKnownType(kind, &spec.Meta{})

	kinds := s.Kinds(&spec.Meta{})
	assert.Len(t, kinds, 1)
	assert.Equal(t, kind, kinds[0])
}
