package primitive

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewBool(t *testing.T) {
	v := NewBool(true)

	assert.Equal(t, KindBool, v.Kind())
	assert.Equal(t, true, v.Interface())
}

func TestBool_Compare(t *testing.T) {
	assert.Equal(t, 0, TRUE.Compare(TRUE))
	assert.Equal(t, 0, FALSE.Compare(FALSE))
	assert.Equal(t, 1, TRUE.Compare(FALSE))
	assert.Equal(t, -1, FALSE.Compare(TRUE))
}

func TestBool_Equal(t *testing.T) {
	assert.True(t, TRUE.Equal(TRUE))
	assert.True(t, FALSE.Equal(FALSE))
	assert.False(t, TRUE.Equal(FALSE))
	assert.False(t, FALSE.Equal(TRUE))
}

func TestBool_Encode(t *testing.T) {
	e := NewBoolEncoder()

	v, err := e.Encode(true)
	assert.NoError(t, err)
	assert.Equal(t, TRUE, v)
}

func TestBool_Decode(t *testing.T) {
	d := NewBoolDecoder()

	var v bool
	err := d.Decode(TRUE, &v)
	assert.NoError(t, err)
	assert.Equal(t, true, v)
}
