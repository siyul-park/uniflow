package types

import (
	"reflect"
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
	enc := encoding.NewEncodeAssembler[any, Value]()
	enc.Add(newStringEncoder())

	t.Run("encoding.TextMarshaler", func(t *testing.T) {
		source := uuid.Must(uuid.NewV7())
		v := NewString(source.String())

		decoded, err := enc.Encode(source)
		assert.NoError(t, err)
		assert.Equal(t, v, decoded)
	})

	t.Run("string", func(t *testing.T) {
		source := faker.Word()
		v := NewString(source)

		decoded, err := enc.Encode(source)
		assert.NoError(t, err)
		assert.Equal(t, v, decoded)
	})
}

func TestString_Decode(t *testing.T) {
	dec := encoding.NewDecodeAssembler[Value, any]()
	dec.Add(newStringDecoder())

	uid := uuid.Must(uuid.NewV7())

	tests := []struct {
		name   string
		source String
		target any
		want   any
	}{
		{"encoding.TextUnmarshaler", NewString(uid.String()), new(uuid.UUID), uid},
		{"bool", NewString("true"), new(bool), true},
		{"float32", NewString("1"), new(float32), float32(1)},
		{"float64", NewString("1"), new(float64), float64(1)},
		{"int", NewString("1"), new(int), 1},
		{"int8", NewString("1"), new(int8), int8(1)},
		{"int16", NewString("1"), new(int16), int16(1)},
		{"int32", NewString("1"), new(int32), int32(1)},
		{"int64", NewString("1"), new(int64), int64(1)},
		{"uint", NewString("1"), new(uint), uint(1)},
		{"uint8", NewString("1"), new(uint8), uint8(1)},
		{"uint16", NewString("1"), new(uint16), uint16(1)},
		{"uint32", NewString("1"), new(uint32), uint32(1)},
		{"uint64", NewString("1"), new(uint64), uint64(1)},
		{"any", NewString("foo"), new(any), "foo"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := dec.Decode(tt.source, tt.target)
			assert.NoError(t, err)
			assert.Equal(t, tt.want, reflect.ValueOf(tt.target).Elem().Interface())
		})
	}
}
