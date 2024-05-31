package object

import (
	"testing"

	"github.com/siyul-park/uniflow/pkg/encoding"

	"github.com/stretchr/testify/assert"
)

func TestBool_NewBool(t *testing.T) {
	assert.Equal(t, True, NewBool(true))
	assert.Equal(t, False, NewBool(false))
}

func TestBool_Bool(t *testing.T) {
	assert.Equal(t, true, True.Bool())
	assert.Equal(t, false, False.Bool())
}

func TestBool_Kind(t *testing.T) {
	assert.Equal(t, KindBool, True.Kind())
}

func TestBool_Hash(t *testing.T) {
	assert.NotEqual(t, True.Hash(), False.Hash())
}

func TestBool_Interface(t *testing.T) {
	assert.Equal(t, true, True.Interface())
	assert.Equal(t, false, False.Interface())
}

func TestBool_Equal(t *testing.T) {
	assert.True(t, True.Equal(True))
	assert.True(t, False.Equal(False))
	assert.False(t, True.Equal(False))
	assert.False(t, False.Equal(True))
}

func TestBool_Compare(t *testing.T) {
	assert.Equal(t, 0, True.Compare(True))
	assert.Equal(t, 0, False.Compare(False))
	assert.Equal(t, -1, False.Compare(True))
	assert.Equal(t, 1, True.Compare(False))
}

func TestBool_Encode(t *testing.T) {
	enc := encoding.NewEncodeAssembler[any, Object]()
	enc.Add(newBoolEncoder())

	source := true
	v := NewBool(source)

	decoded, err := enc.Encode(&source)
	assert.NoError(t, err)
	assert.Equal(t, v, decoded)
}

func TestBool_Decode(t *testing.T) {
	dec := encoding.NewDecodeAssembler[Object, any]()
	dec.Add(newBoolDecoder())

	t.Run("bool", func(t *testing.T) {
		v := NewBool(true)

		var decoded bool
		err := dec.Decode(v, &decoded)
		assert.NoError(t, err)
		assert.Equal(t, true, decoded)
	})

	t.Run("any", func(t *testing.T) {
		v := NewBool(true)

		var decoded any
		err := dec.Decode(v, &decoded)
		assert.NoError(t, err)
		assert.Equal(t, true, decoded)
	})
}
