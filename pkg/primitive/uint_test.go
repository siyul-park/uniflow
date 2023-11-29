package primitive

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestUinteger(t *testing.T) {
	t.Run("Uint", func(t *testing.T) {
		v := NewUint(0)
		assert.Equal(t, KindUint, v.Kind())
		assert.Equal(t, uint(0), v.Interface())
	})

	t.Run("Uint8", func(t *testing.T) {
		v := NewUint8(0)
		assert.Equal(t, KindUint8, v.Kind())
		assert.Equal(t, uint8(0), v.Interface())
	})

	t.Run("Uint16", func(t *testing.T) {
		v := NewUint16(0)
		assert.Equal(t, KindUint16, v.Kind())
		assert.Equal(t, uint16(0), v.Interface())
	})

	t.Run("Uint32", func(t *testing.T) {
		v := NewUint32(0)
		assert.Equal(t, KindUint32, v.Kind())
		assert.Equal(t, uint32(0), v.Interface())
	})

	t.Run("Uint64", func(t *testing.T) {
		v := NewUint64(0)
		assert.Equal(t, KindUint64, v.Kind())
		assert.Equal(t, uint64(0), v.Interface())
	})
}

func TestUinteger_Compare(t *testing.T) {
	t.Run("Uint", func(t *testing.T) {
		assert.Equal(t, 0, NewUint(0).Compare(NewUint(0)))
		assert.Equal(t, 1, NewUint(1).Compare(NewUint(0)))
		assert.Equal(t, -1, NewUint(0).Compare(NewUint(1)))
		assert.Equal(t, 0, NewUint(0).Compare(NewFloat32(0)))
	})

	t.Run("Uint8", func(t *testing.T) {
		assert.Equal(t, 0, NewUint8(0).Compare(NewUint(0)))
		assert.Equal(t, 1, NewUint8(1).Compare(NewUint(0)))
		assert.Equal(t, -1, NewUint8(0).Compare(NewUint(1)))
		assert.Equal(t, 0, NewUint8(0).Compare(NewFloat32(0)))
	})

	t.Run("Uint16", func(t *testing.T) {
		assert.Equal(t, 0, NewUint16(0).Compare(NewUint(0)))
		assert.Equal(t, 1, NewUint16(1).Compare(NewUint(0)))
		assert.Equal(t, -1, NewUint16(0).Compare(NewUint(1)))
		assert.Equal(t, 0, NewUint16(0).Compare(NewFloat32(0)))
	})

	t.Run("Uint32", func(t *testing.T) {
		assert.Equal(t, 0, NewUint32(0).Compare(NewUint(0)))
		assert.Equal(t, 1, NewUint32(1).Compare(NewUint(0)))
		assert.Equal(t, -1, NewUint32(0).Compare(NewUint(1)))
		assert.Equal(t, 0, NewUint32(0).Compare(NewFloat32(0)))
	})

	t.Run("Uint64", func(t *testing.T) {
		assert.Equal(t, 0, NewUint64(0).Compare(NewUint(0)))
		assert.Equal(t, 1, NewUint64(1).Compare(NewUint(0)))
		assert.Equal(t, -1, NewUint64(0).Compare(NewUint(1)))
		assert.Equal(t, 0, NewUint64(0).Compare(NewFloat32(0)))
	})
}

func TestUinteger_EncodeAndDecode(t *testing.T) {
	e := NewUintEncoder()
	d := NewUintDecoder()

	t.Run("Uint", func(t *testing.T) {
		source := uint(1)

		encoded, err := e.Encode(source)
		assert.NoError(t, err)
		assert.Equal(t, Uint(1), encoded)

		var decoded uint
		err = d.Decode(encoded, &decoded)
		assert.NoError(t, err)
		assert.Equal(t, source, decoded)
	})

	t.Run("Uint8", func(t *testing.T) {
		source := uint8(1)

		encoded, err := e.Encode(source)
		assert.NoError(t, err)
		assert.Equal(t, Uint8(1), encoded)

		var decoded uint8
		err = d.Decode(encoded, &decoded)
		assert.NoError(t, err)
		assert.Equal(t, source, decoded)
	})

	t.Run("Uint16", func(t *testing.T) {
		source := uint16(1)

		encoded, err := e.Encode(source)
		assert.NoError(t, err)
		assert.Equal(t, Uint16(1), encoded)

		var decoded uint16
		err = d.Decode(encoded, &decoded)
		assert.NoError(t, err)
		assert.Equal(t, source, decoded)
	})

	t.Run("Uint32", func(t *testing.T) {
		source := uint32(1)

		encoded, err := e.Encode(source)
		assert.NoError(t, err)
		assert.Equal(t, Uint32(1), encoded)

		var decoded uint32
		err = d.Decode(encoded, &decoded)
		assert.NoError(t, err)
		assert.Equal(t, source, decoded)
	})

	t.Run("Uint64", func(t *testing.T) {
		source := uint64(1)

		encoded, err := e.Encode(source)
		assert.NoError(t, err)
		assert.Equal(t, Uint64(1), encoded)

		var decoded uint64
		err = d.Decode(encoded, &decoded)
		assert.NoError(t, err)
		assert.Equal(t, source, decoded)
	})
}
