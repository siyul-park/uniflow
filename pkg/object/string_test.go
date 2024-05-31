package object

import (
	"testing"

	"github.com/gofrs/uuid"
	"github.com/siyul-park/uniflow/pkg/encoding"

	"github.com/go-faker/faker/v4"
	"github.com/stretchr/testify/assert"
)

func TestString_Len(t *testing.T) {
	v := NewString("hello")

	assert.Equal(t, 5, v.Len())
}

func TestString_Get(t *testing.T) {
	v := NewString("hello")

	assert.Equal(t, 'h', v.Get(0))
	assert.Equal(t, 'e', v.Get(1))
	assert.Equal(t, 'l', v.Get(2))
	assert.Equal(t, 'l', v.Get(3))
	assert.Equal(t, 'o', v.Get(4))
	assert.Equal(t, rune(0), v.Get(5))
}

func TestString_String(t *testing.T) {
	v := NewString("hello")

	assert.Equal(t, "hello", v.String())
}

func TestString_Kind(t *testing.T) {
	v := NewString("hello")

	assert.Equal(t, KindString, v.Kind())
}

func TestString_Hash(t *testing.T) {
	v1 := NewString("hello")
	v2 := NewString("world")

	assert.NotEqual(t, v1.Hash(), v2.Hash())
}

func TestString_Interface(t *testing.T) {
	v := NewString("hello")

	assert.Equal(t, "hello", v.Interface())
}

func TestString_Equal(t *testing.T) {
	v1 := NewString("hello")
	v2 := NewString("world")

	assert.True(t, v1.Equal(v1))
	assert.True(t, v2.Equal(v2))
	assert.False(t, v1.Equal(v2))
}

func TestString_Compare(t *testing.T) {
	v1 := NewString("hello")
	v2 := NewString("world")

	assert.Equal(t, 0, v1.Compare(v1))
	assert.Equal(t, 0, v2.Compare(v2))
	assert.Equal(t, -1, v1.Compare(v2))
	assert.Equal(t, 1, v2.Compare(v1))
}

func TestString_Encode(t *testing.T) {
	enc := encoding.NewEncodeAssembler[any, Object]()
	enc.Add(newStringEncoder())

	t.Run("encoding.TextMarshaler", func(t *testing.T) {
		source := uuid.Must(uuid.NewV7())
		v := NewString(source.String())

		decoded, err := enc.Encode(&source)
		assert.NoError(t, err)
		assert.Equal(t, v, decoded)
	})

	t.Run("string", func(t *testing.T) {
		source := faker.Word()
		v := NewString(source)

		decoded, err := enc.Encode(&source)
		assert.NoError(t, err)
		assert.Equal(t, v, decoded)
	})
}

func TestString_Decode(t *testing.T) {
	dec := encoding.NewDecodeAssembler[Object, any]()
	dec.Add(newStringDecoder())

	t.Run("encoding.TextUnmarshaler", func(t *testing.T) {
		source := uuid.Must(uuid.NewV7())
		v := NewString(source.String())

		var decoded uuid.UUID
		err := dec.Decode(v, &decoded)
		assert.NoError(t, err)
		assert.Equal(t, source, decoded)
	})

	t.Run("string", func(t *testing.T) {
		source := faker.Word()
		v := NewString(source)

		var decoded string
		err := dec.Decode(v, &decoded)
		assert.NoError(t, err)
		assert.Equal(t, source, decoded)
	})

	t.Run("any", func(t *testing.T) {
		source := faker.Word()
		v := NewString(source)

		var decoded any
		err := dec.Decode(v, &decoded)
		assert.NoError(t, err)
		assert.Equal(t, source, decoded)
	})
}

func BenchmarkString_Encode(b *testing.B) {
	enc := encoding.NewEncodeAssembler[any, Object]()
	enc.Add(newStringEncoder())

	b.Run("encoding.TextMarshaler", func(b *testing.B) {
		source := uuid.Must(uuid.NewV7())

		for i := 0; i < b.N; i++ {
			enc.Encode(&source)
		}
	})

	b.Run("string", func(b *testing.B) {
		source := faker.Word()

		for i := 0; i < b.N; i++ {
			enc.Encode(&source)
		}
	})
}
