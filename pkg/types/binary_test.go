package types

import (
	"encoding/base64"
	"testing"

	"github.com/gofrs/uuid"
	"github.com/stretchr/testify/require"

	"github.com/siyul-park/uniflow/pkg/encoding"
)

func TestBinary_Len(t *testing.T) {
	v := NewBinary([]byte{0})

	require.Equal(t, 1, v.Len())
}

func TestBinary_Get(t *testing.T) {
	v := NewBinary([]byte{0})

	require.Equal(t, byte(0), v.Get(0))
}

func TestBinary_Bytes(t *testing.T) {
	v := NewBinary([]byte{0})

	require.Equal(t, []byte{0}, v.Bytes())
}

func TestBinary_String(t *testing.T) {
	v := NewBinary([]byte{0})

	require.Equal(t, "AA==", v.String())
}

func TestBinary_Kind(t *testing.T) {
	v := NewBinary([]byte{0})

	require.Equal(t, KindBinary, v.Kind())
}

func TestBinary_Hash(t *testing.T) {
	v1 := NewBinary([]byte{0})
	v2 := NewBinary([]byte{1})

	require.NotEqual(t, v1.Hash(), v2.Hash())
}

func TestBinary_Interface(t *testing.T) {
	v := NewBinary([]byte{0})

	require.Equal(t, []byte{0}, v.Interface())
}

func TestBinary_Equal(t *testing.T) {
	v1 := NewBinary([]byte{0})
	v2 := NewBinary([]byte{1})

	require.True(t, v1.Equal(v1))
	require.True(t, v2.Equal(v2))
	require.False(t, v1.Equal(v2))
}

func TestBinary_Compare(t *testing.T) {
	v1 := NewBinary([]byte{0})
	v2 := NewBinary([]byte{1})

	require.Equal(t, 0, v1.Compare(v1))
	require.Equal(t, 0, v2.Compare(v2))
	require.Equal(t, -1, v1.Compare(v2))
	require.Equal(t, 1, v2.Compare(v1))
}

func TestBinary_MarshalText(t *testing.T) {
	b := NewBinary([]byte{0, 1, 2})
	text, err := b.MarshalText()
	require.NoError(t, err)
	require.Equal(t, "AAEC", string(text))
}

func TestBinary_UnmarshalText(t *testing.T) {
	b := NewBinary(nil)
	err := b.UnmarshalText([]byte("AAEC"))
	require.NoError(t, err)
	require.Equal(t, []byte{0, 1, 2}, b.Bytes())
}

func TestBinary_MarshalBinary(t *testing.T) {
	b := NewBinary([]byte{0, 1, 2})
	data, err := b.MarshalBinary()
	require.NoError(t, err)
	require.Equal(t, []byte{0, 1, 2}, data)
}

func TestBinary_UnmarshalBinary(t *testing.T) {
	b := NewBinary(nil)
	err := b.UnmarshalBinary([]byte{0, 1, 2})
	require.NoError(t, err)
	require.Equal(t, []byte{0, 1, 2}, b.Bytes())
}

func TestBinary_Encode(t *testing.T) {
	enc := encoding.NewEncodeAssembler[any, Value]()
	enc.Add(newBinaryEncoder())

	t.Run("encoding.BinaryMarshaler", func(t *testing.T) {
		source := uuid.Must(uuid.NewV7())
		v := NewBinary(source.Bytes())

		decoded, err := enc.Encode(source)
		require.NoError(t, err)
		require.Equal(t, v, decoded)
	})

	t.Run("slice", func(t *testing.T) {
		source := []byte{0, 1, 2}
		v := NewBinary(source)

		decoded, err := enc.Encode(source)
		require.NoError(t, err)
		require.Equal(t, v, decoded)
	})

	t.Run("array", func(t *testing.T) {
		source := [3]byte{0, 1, 2}
		v := NewBinary(source[:])

		decoded, err := enc.Encode(source)
		require.NoError(t, err)
		require.Equal(t, v, decoded)
	})
}

func TestBinary_Decode(t *testing.T) {
	dec := encoding.NewDecodeAssembler[Value, any]()
	dec.Add(newBinaryDecoder())

	t.Run("encoding.BinaryUnmarshaler", func(t *testing.T) {
		source := uuid.Must(uuid.NewV7())
		v := NewBinary(source.Bytes())

		var decoded uuid.UUID
		err := dec.Decode(v, &decoded)
		require.NoError(t, err)
		require.Equal(t, source, decoded)
	})

	t.Run("slice", func(t *testing.T) {
		source := []byte{0, 1, 2}
		v := NewBinary(source)

		var decoded []byte
		err := dec.Decode(v, &decoded)
		require.NoError(t, err)
		require.Equal(t, source, decoded)
	})

	t.Run("array", func(t *testing.T) {
		source := []byte{0, 1, 2}
		v := NewBinary(source)

		var decoded [3]byte
		err := dec.Decode(v, &decoded)
		require.NoError(t, err)
		require.EqualValues(t, source, decoded)
	})

	t.Run("string", func(t *testing.T) {
		source := []byte{0, 1, 2}
		v := NewBinary(source)

		var decoded string
		err := dec.Decode(v, &decoded)
		require.NoError(t, err)

		d, err := base64.StdEncoding.DecodeString(decoded)
		require.NoError(t, err)
		require.Equal(t, source, d)
	})

	t.Run("any", func(t *testing.T) {
		source := []byte{0, 1, 2}
		v := NewBinary(source)

		var decoded any
		err := dec.Decode(v, &decoded)
		require.NoError(t, err)
		require.Equal(t, source, decoded)
	})
}

func BenchmarkBinary_Encode(b *testing.B) {
	enc := encoding.NewEncodeAssembler[any, Value]()
	enc.Add(newBinaryEncoder())

	b.Run("encoding.BinaryMarshaler", func(b *testing.B) {
		source := uuid.Must(uuid.NewV7())

		for i := 0; i < b.N; i++ {
			enc.Encode(source)
		}
	})

	b.Run("slice", func(b *testing.B) {
		source := []byte{0, 1, 2}

		for i := 0; i < b.N; i++ {
			enc.Encode(source)
		}
	})

	b.Run("array", func(b *testing.B) {
		source := [3]byte{0, 1, 2}

		for i := 0; i < b.N; i++ {
			enc.Encode(source)
		}
	})
}
