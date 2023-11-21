package scheme

import (
	"reflect"
	"testing"

	"github.com/go-faker/faker/v4"
	"github.com/siyul-park/uniflow/pkg/node"
	"github.com/stretchr/testify/assert"
)

func TestScheme_KnownType(t *testing.T) {
	s := New()
	kind := faker.Word()

	s.AddKnownType(kind, &SpecMeta{})

	typ, ok := s.KnownType(kind)
	assert.True(t, ok)
	assert.Equal(t, reflect.TypeOf(&SpecMeta{}), typ)
}

func TestScheme_Codec(t *testing.T) {
	s := New()
	kind := faker.Word()

	c := CodecFunc(func(spec Spec) (node.Node, error) {
		return node.NewOneToOneNode(node.OneToOneNodeConfig{ID: spec.GetID()}), nil
	})

	s.AddCodec(kind, c)

	_, ok := s.Codec(kind)
	assert.True(t, ok)
}

func TestScheme_New(t *testing.T) {
	s := New()
	kind := faker.Word()

	s.AddKnownType(kind, &SpecMeta{})

	spec, ok := s.New(kind)
	assert.True(t, ok)
	assert.IsType(t, spec, &SpecMeta{})
}

func TestScheme_Decode(t *testing.T) {
	s := New()
	kind := faker.Word()

	s.AddKnownType(kind, &SpecMeta{})
	s.AddCodec(kind, CodecFunc(func(spec Spec) (node.Node, error) {
		return node.NewOneToOneNode(node.OneToOneNodeConfig{}), nil
	}))

	n, err := s.Decode(&SpecMeta{})
	assert.NoError(t, err)
	assert.NotNil(t, n)
}

func TestScheme_Kinds(t *testing.T) {
	s := New()
	kind := faker.Word()

	s.AddKnownType(kind, &SpecMeta{})

	kinds := s.Kinds(&SpecMeta{})
	assert.Len(t, kinds, 1)
	assert.Equal(t, kind, kinds[0])
}