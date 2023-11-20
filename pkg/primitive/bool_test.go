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

func TestBool_Hash(t *testing.T) {
	assert.NotEqual(t, TRUE.Hash(), FALSE.Hash())
	assert.Equal(t, TRUE.Hash(), TRUE.Hash())
	assert.Equal(t, FALSE.Hash(), FALSE.Hash())
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
