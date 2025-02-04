package types

import (
	"errors"
	"testing"

	"github.com/siyul-park/uniflow/pkg/encoding"
	"github.com/stretchr/testify/assert"
)

func TestError_NewError(t *testing.T) {
	err := errors.New("test error")
	v := NewError(err)

	assert.NotNil(t, v)
	assert.Equal(t, err, v.value)
}

func TestError_Error(t *testing.T) {
	err := errors.New("test error")
	v := NewError(err)

	assert.Equal(t, "test error", v.Error())
}

func TestError_Kind(t *testing.T) {
	err := errors.New("test error")
	v := NewError(err)

	assert.Equal(t, KindError, v.Kind())
}

func TestError_Hash(t *testing.T) {
	err1 := errors.New("test error 1")
	err2 := errors.New("test error 2")

	v1 := NewError(err1)
	v2 := NewError(err2)

	assert.NotEqual(t, v1.Hash(), v2.Hash())
}

func TestError_Interface(t *testing.T) {
	err := errors.New("test error")
	v := NewError(err)

	assert.Equal(t, err, v.Interface())
}

func TestError_Equal(t *testing.T) {
	err1 := errors.New("test error 1")
	err2 := errors.New("test error 2")

	v1 := NewError(err1)
	v2 := NewError(err2)

	assert.True(t, v1.Equal(v1))
	assert.True(t, v2.Equal(v2))
	assert.False(t, v1.Equal(v2))
}

func TestError_Compare(t *testing.T) {
	err1 := errors.New("test error 1")
	err2 := errors.New("test error 2")

	v1 := NewError(err1)
	v2 := NewError(err2)

	assert.Equal(t, 0, v1.Compare(v1))
	assert.NotEqual(t, 0, v1.Compare(v2))
	assert.NotEqual(t, 0, v2.Compare(v1))
}

func TestError_MarshalText(t *testing.T) {
	err := errors.New("test error")
	v := NewError(err)

	text, err := v.MarshalText()
	assert.NoError(t, err)
	assert.Equal(t, "test error", string(text))
}

func TestError_UnmarshalText(t *testing.T) {
	v := NewError(errors.New("test error"))

	err := v.UnmarshalText([]byte("test error"))
	assert.NoError(t, err)
	assert.Equal(t, "test error", v.Error())
}

func TestError_MarshalBinary(t *testing.T) {
	err := errors.New("test error")
	v := NewError(err)

	data, err := v.MarshalBinary()
	assert.NoError(t, err)
	assert.Equal(t, []byte("test error"), data)
}

func TestError_UnmarshalBinary(t *testing.T) {
	v := NewError(errors.New("test error"))

	err := v.UnmarshalBinary([]byte("test error"))
	assert.NoError(t, err)
	assert.Equal(t, "test error", v.Error())
}

func TestError_Encode(t *testing.T) {
	enc := encoding.NewEncodeAssembler[any, Value]()
	enc.Add(newErrorEncoder())

	err := errors.New("test error")
	v := NewError(err)

	encoded, encodeErr := enc.Encode(err)
	assert.NoError(t, encodeErr)
	assert.Equal(t, v, encoded)
}

func TestError_Decode(t *testing.T) {
	dec := encoding.NewDecodeAssembler[Value, any]()
	dec.Add(newErrorDecoder())

	t.Run("error", func(t *testing.T) {
		err := errors.New("test error")
		v := NewError(err)

		var decoded error
		assert.NoError(t, dec.Decode(v, &decoded))
		assert.Equal(t, err, decoded)
	})

	t.Run("string", func(t *testing.T) {
		err := errors.New("test error")
		v := NewError(err)

		var decoded string
		err = dec.Decode(v, &decoded)
		assert.NoError(t, err)
		assert.Equal(t, "test error", decoded)
	})
}
