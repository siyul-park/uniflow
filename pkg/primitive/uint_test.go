package primitive

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewUint(t *testing.T) {
	t.Run("", func(t *testing.T) {
		v := NewUint(0)

		assert.Equal(t, KindUint, v.Kind())
		assert.Equal(t, uint(0), v.Interface())
	})
	t.Run("8", func(t *testing.T) {
		v := NewUint8(0)

		assert.Equal(t, KindUint8, v.Kind())
		assert.Equal(t, uint8(0), v.Interface())
	})
	t.Run("16", func(t *testing.T) {
		v := NewUint16(0)

		assert.Equal(t, KindUint16, v.Kind())
		assert.Equal(t, uint16(0), v.Interface())
	})
	t.Run("32", func(t *testing.T) {
		v := NewUint32(0)

		assert.Equal(t, KindUint32, v.Kind())
		assert.Equal(t, uint32(0), v.Interface())
	})
	t.Run("64", func(t *testing.T) {
		v := NewUint64(0)

		assert.Equal(t, KindUint64, v.Kind())
		assert.Equal(t, uint64(0), v.Interface())
	})
}

func TestUint_Equal(t *testing.T) {
	t.Run("", func(t *testing.T) {
		assert.True(t, NewInt(0).Equal(NewUint(0)))
		assert.False(t, NewUint(0).Equal(NewUint(1)))
		assert.True(t, NewUint(0).Equal(NewFloat32(0)))
	})
	t.Run("8", func(t *testing.T) {
		assert.True(t, NewUint8(0).Equal(NewUint(0)))
		assert.False(t, NewUint8(0).Equal(NewUint(1)))
		assert.True(t, NewUint8(0).Equal(NewFloat32(0)))
	})
	t.Run("16", func(t *testing.T) {
		assert.True(t, NewUint16(0).Equal(NewUint(0)))
		assert.False(t, NewUint16(0).Equal(NewUint(1)))
		assert.True(t, NewUint16(0).Equal(NewFloat32(0)))
	})
	t.Run("32", func(t *testing.T) {
		assert.True(t, NewUint32(0).Equal(NewUint(0)))
		assert.False(t, NewUint32(0).Equal(NewUint(1)))
		assert.True(t, NewUint32(0).Equal(NewFloat32(0)))
	})
	t.Run("64", func(t *testing.T) {
		assert.True(t, NewUint64(0).Equal(NewUint(0)))
		assert.False(t, NewUint64(0).Equal(NewUint(1)))
		assert.True(t, NewUint64(0).Equal(NewFloat32(0)))
	})
}

func TestUint_Compare(t *testing.T) {
	t.Run("", func(t *testing.T) {
		assert.Equal(t, 0, NewUint(0).Compare(NewUint(0)))
		assert.Equal(t, 1, NewUint(1).Compare(NewUint(0)))
		assert.Equal(t, -1, NewUint(0).Compare(NewUint(1)))
		assert.Equal(t, 0, NewUint(0).Compare(NewFloat32(0)))
	})
	t.Run("8", func(t *testing.T) {
		assert.Equal(t, 0, NewUint8(0).Compare(NewUint(0)))
		assert.Equal(t, 1, NewUint8(1).Compare(NewUint(0)))
		assert.Equal(t, -1, NewUint8(0).Compare(NewUint(1)))
		assert.Equal(t, 0, NewUint8(0).Compare(NewFloat32(0)))
	})
	t.Run("16", func(t *testing.T) {
		assert.Equal(t, 0, NewUint16(0).Compare(NewUint(0)))
		assert.Equal(t, 1, NewUint16(1).Compare(NewUint(0)))
		assert.Equal(t, -1, NewUint16(0).Compare(NewUint(1)))
		assert.Equal(t, 0, NewUint16(0).Compare(NewFloat32(0)))
	})
	t.Run("32", func(t *testing.T) {
		assert.Equal(t, 0, NewUint32(0).Compare(NewUint(0)))
		assert.Equal(t, 1, NewUint32(1).Compare(NewUint(0)))
		assert.Equal(t, -1, NewUint32(0).Compare(NewUint(1)))
		assert.Equal(t, 0, NewUint32(0).Compare(NewFloat32(0)))
	})
	t.Run("64", func(t *testing.T) {
		assert.Equal(t, 0, NewUint64(0).Compare(NewUint(0)))
		assert.Equal(t, 1, NewUint64(1).Compare(NewUint(0)))
		assert.Equal(t, -1, NewUint64(0).Compare(NewUint(1)))
		assert.Equal(t, 0, NewUint64(0).Compare(NewFloat32(0)))
	})
}

func TestUint_Encode(t *testing.T) {
	e := NewUintEncoder()

	t.Run("", func(t *testing.T) {
		v, err := e.Encode(uint(1))
		assert.NoError(t, err)
		assert.Equal(t, NewUint(1), v)
	})
	t.Run("8", func(t *testing.T) {
		v, err := e.Encode(uint8(1))
		assert.NoError(t, err)
		assert.Equal(t, NewUint8(1), v)
	})
	t.Run("16", func(t *testing.T) {
		v, err := e.Encode(uint16(1))
		assert.NoError(t, err)
		assert.Equal(t, NewUint16(1), v)
	})
	t.Run("32", func(t *testing.T) {
		v, err := e.Encode(uint32(1))
		assert.NoError(t, err)
		assert.Equal(t, NewUint32(1), v)
	})
	t.Run("64", func(t *testing.T) {
		v, err := e.Encode(uint64(1))
		assert.NoError(t, err)
		assert.Equal(t, NewUint64(1), v)
	})
}

func TestUint_Decode(t *testing.T) {
	d := NewUintDecoder()

	t.Run("", func(t *testing.T) {
		var v uint
		err := d.Decode(NewUint(1), &v)
		assert.NoError(t, err)
		assert.Equal(t, uint(1), v)
	})
	t.Run("8", func(t *testing.T) {
		var v uint8
		err := d.Decode(NewUint8(1), &v)
		assert.NoError(t, err)
		assert.Equal(t, uint8(1), v)
	})
	t.Run("16", func(t *testing.T) {
		var v uint16
		err := d.Decode(NewUint16(1), &v)
		assert.NoError(t, err)
		assert.Equal(t, uint16(1), v)
	})
	t.Run("32", func(t *testing.T) {
		var v uint32
		err := d.Decode(NewUint32(1), &v)
		assert.NoError(t, err)
		assert.Equal(t, uint32(1), v)
	})
	t.Run("64", func(t *testing.T) {
		var v uint64
		err := d.Decode(NewUint64(1), &v)
		assert.NoError(t, err)
		assert.Equal(t, uint64(1), v)
	})
}
