package event

import (
	"testing"

	"github.com/go-faker/faker/v4"
	"github.com/stretchr/testify/assert"
)

func TestNew(t *testing.T) {
	name := faker.UUIDHyphenated()

	e := New(name, nil)
	assert.Equal(t, name, e.Name())
}

func TestEvent_Get(t *testing.T) {
	e := New(faker.UUIDHyphenated(), nil)

	key := faker.UUIDHyphenated()
	val := faker.UUIDHyphenated()

	res, ok := e.Get(key)
	assert.False(t, ok)
	assert.Nil(t, res)

	e.Set(key, val)

	res, ok = e.Get(key)
	assert.True(t, ok)
	assert.Equal(t, val, res)
}

func TestEvent_Set(t *testing.T) {
	e := New(faker.UUIDHyphenated(), nil)

	key := faker.UUIDHyphenated()
	val1 := faker.UUIDHyphenated()
	val2 := faker.UUIDHyphenated()

	e.Set(key, val1)
	e.Set(key, val2)

	res, ok := e.Get(key)
	assert.True(t, ok)
	assert.Equal(t, val2, res)
}

func TestEvent_Delete(t *testing.T) {
	e := New(faker.UUIDHyphenated(), nil)

	key := faker.UUIDHyphenated()
	val := faker.UUIDHyphenated()

	e.Set(key, val)
	e.Delete(key)

	res, ok := e.Get(key)
	assert.False(t, ok)
	assert.Nil(t, res)
}
