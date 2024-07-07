package types

import (
	"testing"

	"github.com/siyul-park/uniflow/pkg/encoding"
	"github.com/stretchr/testify/assert"
)

func TestBoolean_NewBoolean(t *testing.T) {
	assert.Equal(t, True, NewBoolean(true))
	assert.Equal(t, False, NewBoolean(false))
}

func TestBoolean_Boolean(t *testing.T) {
	assert.Equal(t, true, True.Bool())
	assert.Equal(t, false, False.Bool())
}

func TestBoolean_Kind(t *testing.T) {
	assert.Equal(t, KindBoolean, True.Kind())
}

func TestBoolean_Hash(t *testing.T) {
	assert.NotEqual(t, True.Hash(), False.Hash())
}

func TestBoolean_Interface(t *testing.T) {
	assert.Equal(t, true, True.Interface())
	assert.Equal(t, false, False.Interface())
}

func TestBoolean_Equal(t *testing.T) {
	assert.True(t, True.Equal(True))
	assert.True(t, False.Equal(False))
	assert.False(t, True.Equal(False))
	assert.False(t, False.Equal(True))
}

func TestBoolean_Compare(t *testing.T) {
	assert.Equal(t, 0, True.Compare(True))
	assert.Equal(t, 0, False.Compare(False))
	assert.Equal(t, -1, False.Compare(True))
	assert.Equal(t, 1, True.Compare(False))
}

func TestBoolean_Encode(t *testing.T) {
	enc := encoding.NewEncodeAssembler[any, Value]()
	enc.Add(newBooleanEncoder())

	source := true
	v := NewBoolean(source)

	decoded, err := enc.Encode(source)
	assert.NoError(t, err)
	assert.Equal(t, v, decoded)
}

func TestBoolean_Decode(t *testing.T) {
	dec := encoding.NewDecodeAssembler[Value, any]()
	dec.Add(newBooleanDecoder())

	t.Run("bool", func(t *testing.T) {
		v := NewBoolean(true)

		var decoded bool
		err := dec.Decode(v, &decoded)
		assert.NoError(t, err)
		assert.Equal(t, true, decoded)
	})

	t.Run("any", func(t *testing.T) {
		v := NewBoolean(true)

		var decoded any
		err := dec.Decode(v, &decoded)
		assert.NoError(t, err)
		assert.Equal(t, true, decoded)
	})
}
