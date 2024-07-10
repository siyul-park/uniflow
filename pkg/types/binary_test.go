package types

import (
	"encoding/base64"
	"testing"

	"github.com/gofrs/uuid"
	"github.com/siyul-park/uniflow/pkg/encoding"

	"github.com/stretchr/testify/assert"
)

func TestBinary_Len(t *testing.T) {
	v := NewBinary([]byte{0})

	assert.Equal(t, 1, v.Len())
}

func TestBinary_Get(t *testing.T) {
	v := NewBinary([]byte{0})

	assert.Equal(t, byte(0), v.Get(0))
}

func TestBinary_Bytes(t *testing.T) {
	v := NewBinary([]byte{0})

	assert.Equal(t, []byte{0}, v.Bytes())
}

func TestBinary_Kind(t *testing.T) {
	v := NewBinary([]byte{0})

	assert.Equal(t, KindBinary, v.Kind())
}

func TestBinary_Hash(t *testing.T) {
	v1 := NewBinary([]byte{0})
	v2 := NewBinary([]byte{1})

	assert.NotEqual(t, v1.Hash(), v2.Hash())
}

func TestBinary_Interface(t *testing.T) {
	v := NewBinary([]byte{0})

	assert.Equal(t, []byte{0}, v.Interface())
}

func TestBinary_Equal(t *testing.T) {
	v1 := NewBinary([]byte{0})
	v2 := NewBinary([]byte{1})

	assert.True(t, v1.Equal(v1))
	assert.True(t, v2.Equal(v2))
	assert.False(t, v1.Equal(v2))
}

func TestBinary_Compare(t *testing.T) {
	v1 := NewBinary([]byte{0})
	v2 := NewBinary([]byte{1})

	assert.Equal(t, 0, v1.Compare(v1))
	assert.Equal(t, 0, v2.Compare(v2))
	assert.Equal(t, -1, v1.Compare(v2))
	assert.Equal(t, 1, v2.Compare(v1))
}

func TestBinary_Encode(t *testing.T) {
	enc := encoding.NewEncodeAssembler[any, Value]()
	enc.Add(newBinaryEncoder())

	t.Run("encoding.BinaryMarshaler", func(t *testing.T) {
		source := uuid.Must(uuid.NewV7())
		v := NewBinary(source.Bytes())

		decoded, err := enc.Encode(source)
		assert.NoError(t, err)
		assert.Equal(t, v, decoded)
	})

	t.Run("slice", func(t *testing.T) {
		source := []byte{0, 1, 2}
		v := NewBinary(source)

		decoded, err := enc.Encode(source)
		assert.NoError(t, err)
		assert.Equal(t, v, decoded)
	})

	t.Run("array", func(t *testing.T) {
		source := [3]byte{0, 1, 2}
		v := NewBinary(source[:])

		decoded, err := enc.Encode(source)
		assert.NoError(t, err)
		assert.Equal(t, v, decoded)
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
		assert.NoError(t, err)
		assert.Equal(t, source, decoded)
	})

	t.Run("slice", func(t *testing.T) {
		source := []byte{0, 1, 2}
		v := NewBinary(source)

		var decoded []byte
		err := dec.Decode(v, &decoded)
		assert.NoError(t, err)
		assert.Equal(t, source, decoded)
	})

	t.Run("array", func(t *testing.T) {
		source := []byte{0, 1, 2}
		v := NewBinary(source)

		var decoded [3]byte
		err := dec.Decode(v, &decoded)
		assert.NoError(t, err)
		assert.EqualValues(t, source, decoded)
	})

	t.Run("string", func(t *testing.T) {
		source := []byte{0, 1, 2}
		v := NewBinary(source)

		var decoded string
		err := dec.Decode(v, &decoded)
		assert.NoError(t, err)

		d, err := base64.StdEncoding.DecodeString(decoded)
		assert.NoError(t, err)

		assert.Equal(t, source, d)
	})

	t.Run("any", func(t *testing.T) {
		source := []byte{0, 1, 2}
		v := NewBinary(source)

		var decoded any
		err := dec.Decode(v, &decoded)
		assert.NoError(t, err)
		assert.Equal(t, source, decoded)
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
