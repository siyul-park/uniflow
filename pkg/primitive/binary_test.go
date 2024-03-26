package primitive

import (
	"github.com/gofrs/uuid"
	"github.com/siyul-park/uniflow/pkg/encoding"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewBinary(t *testing.T) {
	v := NewBinary([]byte{0})

	assert.Equal(t, KindBinary, v.Kind())
	assert.Equal(t, []byte{0}, v.Interface())
}

func TestBinary_GetAndLen(t *testing.T) {
	v := NewBinary([]byte{0})

	assert.Equal(t, 1, v.Len())
	assert.Equal(t, byte(0), v.Get(0))
}

func TestBinary_Compare(t *testing.T) {
	v1 := NewBinary([]byte{0})
	v2 := NewBinary([]byte{1})

	assert.Equal(t, 0, v1.Compare(v1))
	assert.Equal(t, -1, v1.Compare(v2))
	assert.Equal(t, 1, v2.Compare(v1))
}

func TestBinary_Encode(t *testing.T) {
	enc := encoding.NewAssembler[*Value, any]()
	enc.Add(newBinaryEncoder())

	t.Run("encoding.BinaryMarshaler", func(t *testing.T) {
		source := uuid.Must(uuid.NewV7())
		v := NewBinary(source.Bytes())

		var decoded Value
		err := enc.Encode(&decoded, &source)
		assert.NoError(t, err)
		assert.Equal(t, v, decoded)
	})

	t.Run("slice", func(t *testing.T) {
		source := []byte{0, 1, 2}
		v := NewBinary(source)

		var decoded Value
		err := enc.Encode(&decoded, &source)
		assert.NoError(t, err)
		assert.Equal(t, v, decoded)
	})

	t.Run("array", func(t *testing.T) {
		source := [3]byte{0, 1, 2}
		v := NewBinary(source[:])

		var decoded Value
		err := enc.Encode(&decoded, &source)
		assert.NoError(t, err)
		assert.Equal(t, v, decoded)
	})
}

func TestBinary_Decode(t *testing.T) {
	dec := encoding.NewAssembler[Value, any]()
	dec.Add(newBinaryDecoder())

	t.Run("encoding.BinaryUnmarshaler", func(t *testing.T) {
		source := uuid.Must(uuid.NewV7())
		v := NewBinary(source.Bytes())

		var decoded uuid.UUID
		err := dec.Encode(v, &decoded)
		assert.NoError(t, err)
		assert.Equal(t, source, decoded)
	})

	t.Run("slice", func(t *testing.T) {
		source := []byte{0, 1, 2}
		v := NewBinary(source)

		var decoded []byte
		err := dec.Encode(v, &decoded)
		assert.NoError(t, err)
		assert.Equal(t, source, decoded)
	})

	t.Run("array", func(t *testing.T) {
		source := []byte{0, 1, 2}
		v := NewBinary(source)

		var decoded [3]byte
		err := dec.Encode(v, &decoded)
		assert.NoError(t, err)
		assert.EqualValues(t, source, decoded)
	})

	t.Run("any", func(t *testing.T) {
		source := []byte{0, 1, 2}
		v := NewBinary(source)

		var decoded any
		err := dec.Encode(v, &decoded)
		assert.NoError(t, err)
		assert.Equal(t, source, decoded)
	})
}

func BenchmarkBinary_Encode(b *testing.B) {
	enc := encoding.NewAssembler[*Value, any]()
	enc.Add(newBinaryEncoder())

	b.Run("encoding.BinaryMarshaler", func(b *testing.B) {
		source := uuid.Must(uuid.NewV7())

		for i := 0; i < b.N; i++ {
			var decoded Value
			_ = enc.Encode(&decoded, &source)
		}
	})

	b.Run("slice", func(b *testing.B) {
		source := []byte{0, 1, 2}

		for i := 0; i < b.N; i++ {
			var decoded Value
			_ = enc.Encode(&decoded, &source)
		}
	})

	b.Run("array", func(b *testing.B) {
		source := [3]byte{0, 1, 2}

		for i := 0; i < b.N; i++ {
			var decoded Value
			_ = enc.Encode(&decoded, &source)
		}
	})
}

func BenchmarkBinary_Decode(b *testing.B) {
	dec := encoding.NewAssembler[Value, any]()
	dec.Add(newBinaryDecoder())

	b.Run("encoding.BinaryUnmarshaler", func(b *testing.B) {
		source := uuid.Must(uuid.NewV7())
		v := NewBinary(source.Bytes())

		for i := 0; i < b.N; i++ {
			var decoded uuid.UUID
			_ = dec.Encode(v, &decoded)
		}
	})

	b.Run("slice", func(b *testing.B) {
		source := []byte{0, 1, 2}
		v := NewBinary(source)

		for i := 0; i < b.N; i++ {
			var decoded []byte
			_ = dec.Encode(v, &decoded)
		}
	})

	b.Run("array", func(b *testing.B) {
		source := []byte{0, 1, 2}
		v := NewBinary(source)

		for i := 0; i < b.N; i++ {
			var decoded [3]byte
			_ = dec.Encode(v, &decoded)
		}
	})

	b.Run("any", func(b *testing.B) {
		source := []byte{0, 1, 2}
		v := NewBinary(source)

		for i := 0; i < b.N; i++ {
			var decoded any
			_ = dec.Encode(v, &decoded)
		}
	})
}
