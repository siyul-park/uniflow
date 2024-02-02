package process

import (
	"testing"

	"github.com/go-faker/faker/v4"
	"github.com/stretchr/testify/assert"
)

func TestShared_Load(t *testing.T) {
	s := newShare()

	key := faker.UUIDHyphenated()
	val := faker.UUIDHyphenated()

	s.Store(key, val)

	res := s.Load(key)
	assert.Equal(t, val, res)
}

func TestShared_Store(t *testing.T) {
	s := newShare()

	key := faker.UUIDHyphenated()
	val := faker.UUIDHyphenated()

	ok := s.Store(key, val)
	assert.True(t, ok)

	ok = s.Store(key, val)
	assert.False(t, ok)

	res := s.Load(key)
	assert.Equal(t, val, res)
}

func TestShared_Delete(t *testing.T) {
	s := newShare()

	key := faker.UUIDHyphenated()
	val := faker.UUIDHyphenated()

	ok := s.Delete(key)
	assert.False(t, ok)

	s.Store(key, val)

	ok = s.Delete(key)
	assert.True(t, ok)

	res := s.Load(key)
	assert.Nil(t, res)
}

func TestShared_Close(t *testing.T) {
	s := newShare()

	key := faker.UUIDHyphenated()
	val := faker.UUIDHyphenated()

	s.Store(key, val)

	s.Close()

	res := s.Load(key)
	assert.Nil(t, res)
}
