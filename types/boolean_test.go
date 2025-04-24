package types

import (
	"testing"

	"github.com/siyul-park/uniflow/encoding"
	"github.com/stretchr/testify/require"
)

func TestBoolean_NewBoolean(t *testing.T) {
	require.Equal(t, True, NewBoolean(true))
	require.Equal(t, False, NewBoolean(false))
}

func TestBoolean_Boolean(t *testing.T) {
	require.Equal(t, true, True.Bool())
	require.Equal(t, false, False.Bool())
}

func TestBoolean_Kind(t *testing.T) {
	require.Equal(t, KindBoolean, True.Kind())
}

func TestBoolean_Hash(t *testing.T) {
	require.NotEqual(t, True.Hash(), False.Hash())
}

func TestBoolean_Interface(t *testing.T) {
	require.Equal(t, true, True.Interface())
	require.Equal(t, false, False.Interface())
}

func TestBoolean_Equal(t *testing.T) {
	require.True(t, True.Equal(True))
	require.True(t, False.Equal(False))
	require.False(t, True.Equal(False))
	require.False(t, False.Equal(True))
}

func TestBoolean_Compare(t *testing.T) {
	require.Equal(t, 0, True.Compare(True))
	require.Equal(t, 0, False.Compare(False))
	require.Equal(t, -1, False.Compare(True))
	require.Equal(t, 1, True.Compare(False))
}

func TestBoolean_MarshalJSON(t *testing.T) {
	t.Run("true", func(t *testing.T) {
		v := NewBoolean(true)

		data, err := v.MarshalJSON()
		require.NoError(t, err)
		require.Equal(t, "true", string(data))
	})

	t.Run("false", func(t *testing.T) {
		v := NewBoolean(false)

		data, err := v.MarshalJSON()
		require.NoError(t, err)
		require.Equal(t, "false", string(data))
	})
}

func TestBoolean_UnmarshalJSON(t *testing.T) {
	t.Run("true", func(t *testing.T) {
		v := NewBoolean(false)

		err := v.UnmarshalJSON([]byte("true"))
		require.NoError(t, err)
		require.Equal(t, True, v)
	})

	t.Run("false", func(t *testing.T) {
		v := NewBoolean(true)

		err := v.UnmarshalJSON([]byte("false"))
		require.NoError(t, err)
		require.Equal(t, False, v)
	})
}

func TestBoolean_Encode(t *testing.T) {
	enc := encoding.NewEncodeAssembler[any, Value]()
	enc.Add(newBooleanEncoder())

	source := true
	v := NewBoolean(source)

	decoded, err := enc.Encode(source)
	require.NoError(t, err)
	require.Equal(t, v, decoded)
}

func TestBoolean_Decode(t *testing.T) {
	dec := encoding.NewDecodeAssembler[Value, any]()
	dec.Add(newBooleanDecoder())

	t.Run("bool", func(t *testing.T) {
		v := NewBoolean(true)

		var decoded bool
		err := dec.Decode(v, &decoded)
		require.NoError(t, err)
		require.Equal(t, true, decoded)
	})

	t.Run("string", func(t *testing.T) {
		v := NewBoolean(true)

		var decoded string
		err := dec.Decode(v, &decoded)
		require.NoError(t, err)
		require.Equal(t, "true", decoded)
	})

	t.Run("any", func(t *testing.T) {
		v := NewBoolean(true)

		var decoded any
		err := dec.Decode(v, &decoded)
		require.NoError(t, err)
		require.Equal(t, true, decoded)
	})
}
