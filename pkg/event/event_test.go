package event

import (
	"testing"

	"github.com/go-faker/faker/v4"
	"github.com/siyul-park/uniflow/pkg/primitive"
	"github.com/stretchr/testify/assert"
)

func TestNew(t *testing.T) {
	topic := faker.Word()
	e := New(topic)
	assert.Equal(t, topic, e.Topic())
}

func TestEvent_SetAndGet(t *testing.T) {
	topic := faker.Word()
	e := New(topic)

	key := faker.UUIDHyphenated()
	val := primitive.NewString(faker.UUIDHyphenated())

	e.Set(key, val)

	res, ok := e.Get(key)
	assert.True(t, ok)
	assert.Equal(t, val, res)
}

func TestEvent_MarshalPrimitive(t *testing.T) {
	topic := faker.Word()

	e := New(topic)
	doc := primitive.NewMap(primitive.NewString(KeyTopic), primitive.NewString(topic))

	res, err := e.MarshalPrimitive()
	assert.NoError(t, err)
	assert.Equal(t, doc, res)
}

func TestEvent_UnmarshalPrimitive(t *testing.T) {
	topic := faker.Word()

	doc := primitive.NewMap(primitive.NewString(KeyTopic), primitive.NewString(topic))

	e := &Event{}

	err := e.UnmarshalPrimitive(doc)
	assert.NoError(t, err)
	assert.Equal(t, topic, e.Topic())
}
