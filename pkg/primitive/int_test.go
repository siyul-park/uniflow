package primitive

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewInteger(t *testing.T) {
	t.Run("Int", func(t *testing.T) {
		v := NewInt(0)
		assert.Equal(t, KindInt, v.Kind())
		assert.Equal(t, int(0), v.Interface())
		assert.Equal(t, int64(0), v.Int())
	})

	t.Run("Int8", func(t *testing.T) {
		v := NewInt8(0)
		assert.Equal(t, KindInt8, v.Kind())
		assert.Equal(t, int8(0), v.Interface())
		assert.Equal(t, int64(0), v.Int())
	})

	t.Run("Int16", func(t *testing.T) {
		v := NewInt16(0)
		assert.Equal(t, KindInt16, v.Kind())
		assert.Equal(t, int16(0), v.Interface())
		assert.Equal(t, int64(0), v.Int())
	})

	t.Run("Int32", func(t *testing.T) {
		v := NewInt32(0)
		assert.Equal(t, KindInt32, v.Kind())
		assert.Equal(t, int32(0), v.Interface())
		assert.Equal(t, int64(0), v.Int())
	})

	t.Run("Int64", func(t *testing.T) {
		v := NewInt64(0)
		assert.Equal(t, KindInt64, v.Kind())
		assert.Equal(t, int64(0), v.Interface())
		assert.Equal(t, int64(0), v.Int())
	})
}

func TestInteger_Compare(t *testing.T) {
	t.Run("Int", func(t *testing.T) {
		assert.Equal(t, 0, NewInt(0).Compare(NewInt(0)))
		assert.Equal(t, 1, NewInt(1).Compare(NewInt(0)))
		assert.Equal(t, -1, NewInt(0).Compare(NewInt(1)))
		assert.Equal(t, 0, NewInt(0).Compare(NewFloat32(0)))
	})

	t.Run("Int8", func(t *testing.T) {
		assert.Equal(t, 0, NewInt8(0).Compare(NewInt(0)))
		assert.Equal(t, 1, NewInt8(1).Compare(NewInt(0)))
		assert.Equal(t, -1, NewInt8(0).Compare(NewInt(1)))
		assert.Equal(t, 0, NewInt8(0).Compare(NewFloat32(0)))
	})

	t.Run("Int16", func(t *testing.T) {
		assert.Equal(t, 0, NewInt16(0).Compare(NewInt(0)))
		assert.Equal(t, 1, NewInt16(1).Compare(NewInt(0)))
		assert.Equal(t, -1, NewInt16(0).Compare(NewInt(1)))
		assert.Equal(t, 0, NewInt16(0).Compare(NewFloat32(0)))
	})

	t.Run("Int32", func(t *testing.T) {
		assert.Equal(t, 0, NewInt32(0).Compare(NewInt(0)))
		assert.Equal(t, 1, NewInt32(1).Compare(NewInt(0)))
		assert.Equal(t, -1, NewInt32(0).Compare(NewInt(1)))
		assert.Equal(t, 0, NewInt32(0).Compare(NewFloat32(0)))
	})

	t.Run("Int64", func(t *testing.T) {
		assert.Equal(t, 0, NewInt64(0).Compare(NewInt(0)))
		assert.Equal(t, 1, NewInt64(1).Compare(NewInt(0)))
		assert.Equal(t, -1, NewInt64(0).Compare(NewInt(1)))
		assert.Equal(t, 0, NewInt64(0).Compare(NewFloat32(0)))
	})
}

func TestInteger_EncodeAndDecode(t *testing.T) {
	e := newIntEncoder()
	d := newIntDecoder()

	t.Run("Int", func(t *testing.T) {
		source := int(1)

		encoded, err := e.Encode(source)
		assert.NoError(t, err)
		assert.Equal(t, Int(1), encoded)

		var decoded int
		err = d.Decode(encoded, &decoded)
		assert.NoError(t, err)
		assert.Equal(t, source, decoded)
	})

	t.Run("Int8", func(t *testing.T) {
		source := int8(1)

		encoded, err := e.Encode(source)
		assert.NoError(t, err)
		assert.Equal(t, Int8(1), encoded)

		var decoded int8
		err = d.Decode(encoded, &decoded)
		assert.NoError(t, err)
		assert.Equal(t, source, decoded)
	})

	t.Run("Int16", func(t *testing.T) {
		source := int16(1)

		encoded, err := e.Encode(source)
		assert.NoError(t, err)
		assert.Equal(t, Int16(1), encoded)

		var decoded int16
		err = d.Decode(encoded, &decoded)
		assert.NoError(t, err)
		assert.Equal(t, source, decoded)
	})

	t.Run("Int32", func(t *testing.T) {
		source := int32(1)

		encoded, err := e.Encode(source)
		assert.NoError(t, err)
		assert.Equal(t, Int32(1), encoded)

		var decoded int32
		err = d.Decode(encoded, &decoded)
		assert.NoError(t, err)
		assert.Equal(t, source, decoded)
	})

	t.Run("Int64", func(t *testing.T) {
		source := int64(1)

		encoded, err := e.Encode(source)
		assert.NoError(t, err)
		assert.Equal(t, Int64(1), encoded)

		var decoded int64
		err = d.Decode(encoded, &decoded)
		assert.NoError(t, err)
		assert.Equal(t, source, decoded)
	})
}
