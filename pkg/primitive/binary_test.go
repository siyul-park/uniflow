package primitive

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewBinary(t *testing.T) {
	v := NewBinary([]byte{0})

	assert.Equal(t, KindBinary, v.Kind())
	assert.Equal(t, []byte{0}, v.Interface())
}

func TestBinary_Hash(t *testing.T) {
	assert.NotEqual(t, NewBinary([]byte{0}).Hash(), NewBinary([]byte{1}).Hash())
	assert.Equal(t, NewBinary(nil).Hash(), NewBinary(nil).Hash())
	assert.Equal(t, NewBinary([]byte{0}).Hash(), NewBinary([]byte{0}).Hash())
}

func TestBinary_Get(t *testing.T) {
	v := NewBinary([]byte{0})

	assert.Equal(t, 1, v.Len())
	assert.Equal(t, byte(0), v.Get(0))
}

func TestBinary_Encode(t *testing.T) {
	e := NewBinaryEncoder()

	v, err := e.Encode([]byte{0})
	assert.NoError(t, err)
	assert.Equal(t, NewBinary([]byte{0}), v)
}

func TestBinary_Decode(t *testing.T) {
	d := NewBinaryDecoder()

	var v []byte
	err := d.Decode(NewBinary([]byte{0}), &v)
	assert.NoError(t, err)
	assert.Equal(t, []byte{0}, v)
}
