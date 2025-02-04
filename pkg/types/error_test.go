package types

import (
	"errors"
	"testing"

	"github.com/siyul-park/uniflow/pkg/encoding"
	"github.com/stretchr/testify/assert"
)

func TestError_NewError(t *testing.T) {
	source := errors.New("test error")
	v := NewError(source)

	assert.NotNil(t, v)
	assert.Equal(t, source, v.value)
}

func TestError_Error(t *testing.T) {
	source := errors.New("test error")
	v := NewError(source)

	assert.Equal(t, "test error", v.Error())
}

func TestError_Kind(t *testing.T) {
	source := errors.New("test error")
	v := NewError(source)

	assert.Equal(t, KindError, v.Kind())
}

func TestError_Hash(t *testing.T) {
	source1 := errors.New("test error 1")
	source2 := errors.New("test error 2")

	v1 := NewError(source1)
	v2 := NewError(source2)

	assert.NotEqual(t, v1.Hash(), v2.Hash())
}

func TestError_Interface(t *testing.T) {
	err := errors.New("test error")
	v := NewError(err)

	assert.Equal(t, err, v.Interface())
}

func TestError_Equal(t *testing.T) {
	source1 := errors.New("test error 1")
	source2 := errors.New("test error 2")

	v1 := NewError(source1)
	v2 := NewError(source2)

	assert.True(t, v1.Equal(v1))
	assert.True(t, v2.Equal(v2))
	assert.False(t, v1.Equal(v2))
}

func TestError_Compare(t *testing.T) {
	source1 := errors.New("test error 1")
	source2 := errors.New("test error 2")

	v1 := NewError(source1)
	v2 := NewError(source2)

	assert.Equal(t, 0, v1.Compare(v1))
	assert.NotEqual(t, 0, v1.Compare(v2))
	assert.NotEqual(t, 0, v2.Compare(v1))
}

func TestError_MarshalText(t *testing.T) {
	source := errors.New("test error")
	v := NewError(source)

	text, source := v.MarshalText()
	assert.NoError(t, source)
	assert.Equal(t, "test error", string(text))
}

func TestError_UnmarshalText(t *testing.T) {
	v := NewError(errors.New("test error"))

	err := v.UnmarshalText([]byte("test error"))
	assert.NoError(t, err)
	assert.Equal(t, "test error", v.Error())
}

func TestError_MarshalBinary(t *testing.T) {
	source := errors.New("test error")
	v := NewError(source)

	data, source := v.MarshalBinary()
	assert.NoError(t, source)
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

	source := errors.New("test error")
	v := NewError(source)

	encoded, err := enc.Encode(source)
	assert.NoError(t, err)
	assert.Equal(t, v, encoded)
}

func TestError_Decode(t *testing.T) {
	dec := encoding.NewDecodeAssembler[Value, any]()
	dec.Add(newErrorDecoder())

	t.Run("encoding.BinaryUnmarshaler", func(t *testing.T) {
		source := errors.New("test error")
		v := NewError(source)

		decoded := NewBuffer(nil)
		err := dec.Decode(v, decoded)
		assert.NoError(t, err)

		data, err := decoded.Bytes()
		assert.NoError(t, err)
		assert.Equal(t, source.Error(), string(data))
	})

	t.Run("encoding.TextUnmarshaler", func(t *testing.T) {
		source := errors.New("test error")
		v := NewError(source)

		decoded := NewString("")
		err := dec.Decode(v, &decoded)
		assert.NoError(t, err)
		assert.Equal(t, source.Error(), decoded.String())
	})

	t.Run("error", func(t *testing.T) {
		source := errors.New("test error")
		v := NewError(source)

		var decoded error
		assert.NoError(t, dec.Decode(v, &decoded))
		assert.Equal(t, source, decoded)
	})

	t.Run("string", func(t *testing.T) {
		source := errors.New("test error")
		v := NewError(source)

		var decoded string
		source = dec.Decode(v, &decoded)
		assert.NoError(t, source)
		assert.Equal(t, "test error", decoded)
	})
}
