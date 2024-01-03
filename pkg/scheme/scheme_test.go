package scheme

import (
	"reflect"
	"testing"

	"github.com/go-faker/faker/v4"
	"github.com/oklog/ulid/v2"
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

func TestScheme_NewSpec(t *testing.T) {
	s := New()
	kind := faker.UUIDHyphenated()

	s.AddKnownType(kind, &SpecMeta{})

	spec, ok := s.NewSpec(kind)
	assert.True(t, ok)
	assert.IsType(t, spec, &SpecMeta{})
}

func TestScheme_NewSpecWithDoc(t *testing.T) {
	s := New()
	kind := faker.UUIDHyphenated()

	s.AddKnownType(kind, &SpecMeta{})

	u := NewUnstructured(nil)
	spec := &SpecMeta{
		ID:   ulid.Make(),
		Kind: faker.UUIDHyphenated(),
	}

	_ = u.Marshal(spec)

	r, err := s.NewSpecWithDoc(u.Doc())
	assert.NoError(t, err)
	assert.NotNil(t, r, spec)
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
