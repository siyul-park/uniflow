package types

import (
	"bytes"
	"io"
	"strings"
	"testing"

	"github.com/gofrs/uuid"
	"github.com/stretchr/testify/require"

	"github.com/siyul-park/uniflow/internal/encoding"
)

func TestBuffer_Read(t *testing.T) {
	r := strings.NewReader("test")
	b := NewBuffer(r)
	p := make([]byte, 4)
	n, err := b.Read(p)
	require.NoError(t, err)
	require.Equal(t, 4, n)
	require.Equal(t, "test", string(p))
}

func TestBuffer_Bytes(t *testing.T) {
	r := strings.NewReader("test")
	b := NewBuffer(r)
	p, err := b.Bytes()
	require.NoError(t, err)
	require.Equal(t, "test", string(p))
}

func TestBuffer_Close(t *testing.T) {
	r := strings.NewReader("test")
	b := NewBuffer(r)
	require.NoError(t, b.Close())
}

func TestBuffer_Kind(t *testing.T) {
	r := strings.NewReader("test")
	b := NewBuffer(r)
	require.Equal(t, KindBuffer, b.Kind())
}

func TestBuffer_Hash(t *testing.T) {
	r1 := strings.NewReader("test1")
	r2 := strings.NewReader("test2")
	b1 := NewBuffer(r1)
	b2 := NewBuffer(r2)
	require.NotEqual(t, b1.Hash(), b2.Hash())
}

func TestBuffer_Interface(t *testing.T) {
	r := strings.NewReader("test")
	b := NewBuffer(r)
	require.Equal(t, r, b.Interface())
}

func TestBuffer_Equal(t *testing.T) {
	r1 := strings.NewReader("test1")
	r2 := strings.NewReader("test2")
	b1 := NewBuffer(r1)
	b2 := NewBuffer(r2)
	require.True(t, b1.Equal(b1))
	require.False(t, b1.Equal(b2))
}

func TestBuffer_Compare(t *testing.T) {
	r1 := strings.NewReader("test1")
	r2 := strings.NewReader("test2")
	b1 := NewBuffer(r1)
	b2 := NewBuffer(r2)
	require.Equal(t, 0, b1.Compare(b1))
	require.NotEqual(t, 0, b1.Compare(b2))
}

func TestBuffer_MarshalBinary(t *testing.T) {
	b := NewBuffer(strings.NewReader("test"))
	data, err := b.MarshalBinary()
	require.NoError(t, err)
	require.Equal(t, []byte("test"), data)
}

func TestBuffer_UnmarshalBinary(t *testing.T) {
	b := NewBuffer(nil)
	err := b.UnmarshalBinary([]byte("test"))
	require.NoError(t, err)

	data, err := b.Bytes()
	require.NoError(t, err)
	require.Equal(t, []byte("test"), data)
}

func TestBuffer_Encode(t *testing.T) {
	enc := encoding.NewEncodeAssembler[any, Value]()
	enc.Add(newBufferEncoder())

	t.Run("io.Reader", func(t *testing.T) {
		source := strings.NewReader("test")
		v := NewBuffer(source)

		decoded, err := enc.Encode(source)
		require.NoError(t, err)
		require.Equal(t, v, decoded)
	})
}

func TestBuffer_Decode(t *testing.T) {
	dec := encoding.NewDecodeAssembler[Value, any]()
	dec.Add(newBufferDecoder())

	t.Run("encoding.BinaryUnmarshaler", func(t *testing.T) {
		source := uuid.Must(uuid.NewV7())
		v := NewBuffer(bytes.NewBuffer(source.Bytes()))

		var decoded uuid.UUID
		err := dec.Decode(v, &decoded)
		require.NoError(t, err)
		require.Equal(t, source, decoded)
	})

	t.Run("encoding.TextUnmarshaler", func(t *testing.T) {
		source := uuid.Must(uuid.NewV7())
		v := NewBuffer(strings.NewReader(source.String()))

		decoded := NewString("")
		err := dec.Decode(v, &decoded)
		require.NoError(t, err)
		require.Equal(t, source.String(), decoded.String())
	})

	t.Run("io.Reader", func(t *testing.T) {
		source := strings.NewReader("test")
		v := NewBuffer(source)

		var decoded io.Reader
		err := dec.Decode(v, &decoded)
		require.NoError(t, err)
		require.Equal(t, source, decoded)
	})

	t.Run("slice", func(t *testing.T) {
		source := strings.NewReader("test")
		v := NewBuffer(source)

		var decoded []byte
		err := dec.Decode(v, &decoded)
		require.NoError(t, err)
		require.Equal(t, []byte("test"), decoded)
	})

	t.Run("array", func(t *testing.T) {
		source := strings.NewReader("test")
		v := NewBuffer(source)

		var decoded [3]byte
		err := dec.Decode(v, &decoded)
		require.NoError(t, err)
		require.EqualValues(t, []byte("test"), decoded)
	})

	t.Run("string", func(t *testing.T) {
		source := strings.NewReader("test")
		v := NewBuffer(source)

		var decoded string
		err := dec.Decode(v, &decoded)
		require.NoError(t, err)
		require.Equal(t, "test", decoded)
	})

	t.Run("any", func(t *testing.T) {
		source := strings.NewReader("test")
		v := NewBuffer(source)

		var decoded any
		err := dec.Decode(v, &decoded)
		require.NoError(t, err)
		require.Equal(t, source, decoded)
	})
}
