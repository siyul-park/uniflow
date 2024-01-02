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

func TestBinary_GetAndLen(t *testing.T) {
	v := NewBinary([]byte{0})

	assert.Equal(t, 1, v.Len())
	assert.Equal(t, byte(0), v.Get(0))
}

func TestBinary_Compare(t *testing.T) {
	v1 := NewBinary([]byte{0})
	v2 := NewBinary([]byte{1})

	assert.Equal(t, 0, v1.Compare(v1))
	assert.Equal(t, -1, v1.Compare(v2))
	assert.Equal(t, 1, v2.Compare(v1))
}

func TestBinary_EncodeAndDecode(t *testing.T) {
	e := newBinaryEncoder()
	d := newBinaryDecoder()

	source := []byte{0}

	encoded, err := e.Encode(source)
	assert.NoError(t, err)
	assert.Equal(t, NewBinary([]byte{0}), encoded)

	var decoded []byte
	err = d.Decode(encoded, &decoded)
	assert.NoError(t, err)
	assert.Equal(t, source, decoded)
}
