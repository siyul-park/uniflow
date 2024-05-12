package event

import (
	"testing"

	"github.com/go-faker/faker/v4"
	"github.com/siyul-park/uniflow/pkg/primitive"
	"github.com/stretchr/testify/assert"
)

func TestEvent_SetAndGet(t *testing.T) {
	e := New()

	key := faker.UUIDHyphenated()
	val := primitive.NewString(faker.UUIDHyphenated())

	e.Set(key, val)

	res := e.Get(key)
	assert.Equal(t, val, res)
}

func TestEvent_MarshalPrimitive(t *testing.T) {
	key := faker.UUIDHyphenated()
	val := primitive.NewString(faker.UUIDHyphenated())

	e := New()
	e.Set(key, val)

	doc := primitive.NewMap(primitive.NewString(key), val)

	res, err := e.MarshalPrimitive()
	assert.NoError(t, err)
	assert.Equal(t, doc, res)
}

func TestEvent_UnmarshalPrimitive(t *testing.T) {
	key := faker.UUIDHyphenated()
	val := primitive.NewString(faker.UUIDHyphenated())

	doc := primitive.NewMap(primitive.NewString(key), val)

	e := &Event{}

	err := e.UnmarshalPrimitive(doc)
	assert.NoError(t, err)
	assert.Equal(t, val, e.Get(key))
}
