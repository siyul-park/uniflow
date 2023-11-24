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
func TestString_Get(t *testing.T) {
	v := NewString("A")

	assert.Equal(t, 1, v.Len())
	assert.Equal(t, rune('A'), v.Get(0))
}

func TestString_Compare(t *testing.T) {
	assert.Equal(t, 0, NewString("A").Compare(NewString("A")))
	assert.Equal(t, 1, NewString("a").Compare(NewString("A")))
	assert.Equal(t, -1, NewString("A").Compare(NewString("a")))
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
