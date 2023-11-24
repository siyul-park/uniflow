package util

import (
	"github.com/go-faker/faker/v4"
	"github.com/stretchr/testify/assert"
	"testing"
)

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
