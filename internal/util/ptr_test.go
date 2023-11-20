package util

import (
	"github.com/go-faker/faker/v4"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestIsNil(t *testing.T) {
	assert.True(t, IsNil(nil))
	assert.False(t, IsNil(1))

	type animal interface{}
	type dog struct{}

	assert.False(t, IsNil(dog{}))

	var d *dog = nil
	var a animal = d
	assert.True(t, IsNil(a))
	assert.Nil(t, d)
}

func TestPtr(t *testing.T) {
	value := faker.UUIDHyphenated()
	assert.Equal(t, value, *Ptr(value))
}

func TestUnPtr(t *testing.T) {
	var nilPtr *string
	assert.Equal(t, "", UnPtr(nilPtr))

	value := faker.UUIDHyphenated()
	ptr := &value
	assert.Equal(t, value, UnPtr(ptr))
}

func TestPtrTo(t *testing.T) {
	assert.Nil(t, PtrTo[int, int](nil, func(s int) int { return s + 1 }))
	assert.Equal(t, UnPtr(PtrTo[int, int](Ptr(1), func(s int) int { return s + 1 })), 2)
}
