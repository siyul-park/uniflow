package primitive

import (
	"testing"

	"github.com/go-faker/faker/v4"
	"github.com/stretchr/testify/assert"
)

func TestNewString(t *testing.T) {
	raw := faker.Word()
	v := NewString(raw)

	assert.Equal(t, KindString, v.Kind())
	assert.Equal(t, raw, v.Interface())
}

func TestString_Hash(t *testing.T) {
	assert.NotEqual(t, NewString("A").Hash(), NewString("B").Hash())
	assert.Equal(t, NewString("").Hash(), NewString("").Hash())
	assert.Equal(t, NewString("A").Hash(), NewString("A").Hash())
}

func TestString_Get(t *testing.T) {
	v := NewString("A")

	assert.Equal(t, 1, v.Len())
	assert.Equal(t, rune('A'), v.Get(0))
}

func TestString_Encode(t *testing.T) {
	e := NewStringEncoder()

	v, err := e.Encode("A")
	assert.NoError(t, err)
	assert.Equal(t, NewString("A"), v)
}

func TestString_Decode(t *testing.T) {
	d := NewStringDecoder()

	var v string
	err := d.Decode(NewString("A"), &v)
	assert.NoError(t, err)
	assert.Equal(t, "A", v)
}
