package template

import (
	"fmt"
	"testing"

	"github.com/go-faker/faker/v4"
	"github.com/stretchr/testify/assert"
)

func TestValues_GetAndSetAndDelete(t *testing.T) {
	vl := NewValues()

	key1 := faker.Word()
	key2 := faker.Word()

	key := fmt.Sprintf("%s.%s", key1, key2)
	value := faker.Word()

	vl.Set(key, value)

	r1, ok := vl.Get(key)
	assert.True(t, ok)
	assert.Equal(t, value, r1)

	_, ok = vl.Get(key1)
	assert.True(t, ok)

	vl.Delete(key)

	r2, ok := vl.Get(key)
	assert.False(t, ok)
	assert.Nil(t, r2)

	_, ok = vl.Get(key1)
	assert.False(t, ok)
}
