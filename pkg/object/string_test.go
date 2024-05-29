package object

import (
	"testing"

	"github.com/gofrs/uuid"
	"github.com/siyul-park/uniflow/pkg/encoding"

	"github.com/go-faker/faker/v4"
	"github.com/stretchr/testify/assert"
)

func TestNewString(t *testing.T) {
	raw := faker.UUIDHyphenated()
	v := NewString(raw)

	assert.Equal(t, KindString, v.Kind())
	assert.Equal(t, raw, v.Interface())
}
func TestString_Get(t *testing.T) {
	v := NewString("A")

	assert.Equal(t, 1, v.Len())
	assert.Equal(t, rune('A'), v.Get(0))
}

func TestString_Compare(t *testing.T) {
	assert.Equal(t, 0, NewString("A").Compare(NewString("A")))
	assert.Equal(t, 1, NewString("a").Compare(NewString("A")))
	assert.Equal(t, -1, NewString("A").Compare(NewString("a")))
}

func TestString_Encode(t *testing.T) {
	enc := encoding.NewAssembler[*Object, any]()
	enc.Add(newStringEncoder())

	t.Run("encoding.TextMarshaler", func(t *testing.T) {
		source := uuid.Must(uuid.NewV7())
		v := NewString(source.String())

		var decoded Object
		err := enc.Encode(&decoded, &source)
		assert.NoError(t, err)
		assert.Equal(t, v, decoded)
	})

	t.Run("string", func(t *testing.T) {
		source := faker.Word()
		v := NewString(source)

		var decoded Object
		err := enc.Encode(&decoded, &source)
		assert.NoError(t, err)
		assert.Equal(t, v, decoded)
	})
}

func TestString_Decode(t *testing.T) {
	dec := encoding.NewAssembler[Object, any]()
	dec.Add(newStringDecoder())

	t.Run("encoding.TextUnmarshaler", func(t *testing.T) {
		source := uuid.Must(uuid.NewV7())
		v := NewString(source.String())

		var decoded uuid.UUID
		err := dec.Encode(v, &decoded)
		assert.NoError(t, err)
		assert.Equal(t, source, decoded)
	})

	t.Run("string", func(t *testing.T) {
		source := faker.Word()
		v := NewString(source)

		var decoded string
		err := dec.Encode(v, &decoded)
		assert.NoError(t, err)
		assert.Equal(t, source, decoded)
	})

	t.Run("any", func(t *testing.T) {
		source := faker.Word()
		v := NewString(source)

		var decoded any
		err := dec.Encode(v, &decoded)
		assert.NoError(t, err)
		assert.Equal(t, source, decoded)
	})
}

func BenchmarkString_Encode(b *testing.B) {
	enc := encoding.NewAssembler[*Object, any]()
	enc.Add(newStringEncoder())

	b.Run("encoding.TextMarshaler", func(b *testing.B) {
		source := uuid.Must(uuid.NewV7())

		for i := 0; i < b.N; i++ {
			var decoded Object
			_ = enc.Encode(&decoded, &source)
		}
	})

	b.Run("string", func(b *testing.B) {
		source := faker.Word()

		for i := 0; i < b.N; i++ {
			var decoded Object
			_ = enc.Encode(&decoded, &source)
		}
	})
}
