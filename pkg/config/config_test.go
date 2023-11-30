package config

import (
	"fmt"
	"testing"

	"github.com/go-faker/faker/v4"
	"github.com/stretchr/testify/assert"
)

func TestConfig_GetAndSetAndDelete(t *testing.T) {
	c := New()

	key1 := faker.Word()
	key2 := faker.Word()

	key := fmt.Sprintf("%s.%s", key1, key2)
	value := faker.Word()

	c.Set(key, value)

	r1, ok := c.Get(key)
	assert.True(t, ok)
	assert.Equal(t, value, r1)

	_, ok = c.Get(key1)
	assert.True(t, ok)

	c.Delete(key)

	r2, ok := c.Get(key)
	assert.False(t, ok)
	assert.Nil(t, r2)

	_, ok = c.Get(key1)
	assert.False(t, ok)
}
