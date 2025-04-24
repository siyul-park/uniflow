package types

import (
	"errors"
	"testing"

	"github.com/siyul-park/uniflow/encoding"
	"github.com/stretchr/testify/require"
)

func TestError_NewError(t *testing.T) {
	source := errors.New("test error")
	v := NewError(source)

	require.NotNil(t, v)
	require.Equal(t, source, v.value)
}

func TestError_Error(t *testing.T) {
	source := errors.New("test error")
	v := NewError(source)

	require.Equal(t, "test error", v.Error())
}

func TestError_Kind(t *testing.T) {
	source := errors.New("test error")
	v := NewError(source)

	require.Equal(t, KindError, v.Kind())
}

func TestError_Hash(t *testing.T) {
	source1 := errors.New("test error 1")
	source2 := errors.New("test error 2")

	v1 := NewError(source1)
	v2 := NewError(source2)

	require.NotEqual(t, v1.Hash(), v2.Hash())
}

func TestError_Interface(t *testing.T) {
	err := errors.New("test error")
	v := NewError(err)

	require.Equal(t, err, v.Interface())
}

func TestError_Equal(t *testing.T) {
	source1 := errors.New("test error 1")
	source2 := errors.New("test error 2")

	v1 := NewError(source1)
	v2 := NewError(source2)

	require.True(t, v1.Equal(v1))
	require.True(t, v2.Equal(v2))
	require.False(t, v1.Equal(v2))
}

func TestError_Compare(t *testing.T) {
	source1 := errors.New("test error 1")
	source2 := errors.New("test error 2")

	v1 := NewError(source1)
	v2 := NewError(source2)

	require.Equal(t, 0, v1.Compare(v1))
	require.NotEqual(t, 0, v1.Compare(v2))
	require.NotEqual(t, 0, v2.Compare(v1))
}

func TestError_MarshalText(t *testing.T) {
	source := errors.New("test error")
	v := NewError(source)

	text, source := v.MarshalText()
	require.NoError(t, source)
	require.Equal(t, "test error", string(text))
}

func TestError_UnmarshalText(t *testing.T) {
	v := NewError(errors.New("test error"))

	err := v.UnmarshalText([]byte("test error"))
	require.NoError(t, err)
	require.Equal(t, "test error", v.Error())
}

func TestError_MarshalBinary(t *testing.T) {
	source := errors.New("test error")
	v := NewError(source)

	data, source := v.MarshalBinary()
	require.NoError(t, source)
	require.Equal(t, []byte("test error"), data)
}

func TestError_UnmarshalBinary(t *testing.T) {
	v := NewError(errors.New("test error"))

	err := v.UnmarshalBinary([]byte("test error"))
	require.NoError(t, err)
	require.Equal(t, "test error", v.Error())
}

func TestError_Encode(t *testing.T) {
	enc := encoding.NewEncodeAssembler[any, Value]()
	enc.Add(newErrorEncoder())

	source := errors.New("test error")
	v := NewError(source)

	encoded, err := enc.Encode(source)
	require.NoError(t, err)
	require.Equal(t, v, encoded)
}

func TestError_Decode(t *testing.T) {
	dec := encoding.NewDecodeAssembler[Value, any]()
	dec.Add(newErrorDecoder())

	t.Run("encoding.BinaryUnmarshaler", func(t *testing.T) {
		source := errors.New("test error")
		v := NewError(source)

		decoded := NewBuffer(nil)
		require.NoError(t, dec.Decode(v, decoded))

		data, err := decoded.Bytes()
		require.NoError(t, err)
		require.Equal(t, source.Error(), string(data))
	})

	t.Run("encoding.TextUnmarshaler", func(t *testing.T) {
		source := errors.New("test error")
		v := NewError(source)

		decoded := NewString("")
		require.NoError(t, dec.Decode(v, &decoded))
		require.Equal(t, source.Error(), decoded.String())
	})

	t.Run("error", func(t *testing.T) {
		source := errors.New("test error")
		v := NewError(source)

		var decoded error
		require.NoError(t, dec.Decode(v, &decoded))
		require.Equal(t, source, decoded)
	})

	t.Run("string", func(t *testing.T) {
		source := errors.New("test error")
		v := NewError(source)

		var decoded string
		require.NoError(t, dec.Decode(v, &decoded))
		require.Equal(t, "test error", decoded)
	})

	t.Run("any", func(t *testing.T) {
		source := errors.New("test error")
		v := NewError(source)

		var decoded any
		require.NoError(t, dec.Decode(v, &decoded))
		require.Equal(t, source, decoded)
	})
}
