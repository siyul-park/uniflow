package primitive

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewBool(t *testing.T) {
	v := NewBool(true)

	assert.Equal(t, KindBool, v.Kind())
	assert.Equal(t, true, v.Interface())
	assert.Equal(t, true, v.Bool())
}

func TestBool_Compare(t *testing.T) {
	assert.Equal(t, 0, TRUE.Compare(TRUE))
	assert.Equal(t, 0, FALSE.Compare(FALSE))
	assert.Equal(t, 1, TRUE.Compare(FALSE))
	assert.Equal(t, -1, FALSE.Compare(TRUE))
	assert.Equal(t, 1, TRUE.Compare(FALSE))
	assert.Equal(t, -1, FALSE.Compare(TRUE))
}

func TestBool_EncodeAndDecode(t *testing.T) {
	e := newBoolEncoder()
	d := newBoolDecoder()

	source := true

	encoded, err := e.Encode(source)
	assert.NoError(t, err)
	assert.Equal(t, TRUE, encoded)

	var decoded bool
	err = d.Decode(encoded, &decoded)
	assert.NoError(t, err)
	assert.Equal(t, source, decoded)
}
