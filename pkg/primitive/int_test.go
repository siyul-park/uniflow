package primitive

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewInt(t *testing.T) {
	t.Run("", func(t *testing.T) {
		v := NewInt(0)

		assert.Equal(t, KindInt, v.Kind())
		assert.Equal(t, int(0), v.Interface())
	})
	t.Run("8", func(t *testing.T) {
		v := NewInt8(0)

		assert.Equal(t, KindInt8, v.Kind())
		assert.Equal(t, int8(0), v.Interface())
	})
	t.Run("16", func(t *testing.T) {
		v := NewInt16(0)

		assert.Equal(t, KindInt16, v.Kind())
		assert.Equal(t, int16(0), v.Interface())
	})
	t.Run("32", func(t *testing.T) {
		v := NewInt32(0)

		assert.Equal(t, KindInt32, v.Kind())
		assert.Equal(t, int32(0), v.Interface())
	})
	t.Run("64", func(t *testing.T) {
		v := NewInt64(0)

		assert.Equal(t, KindInt64, v.Kind())
		assert.Equal(t, int64(0), v.Interface())
	})
}

func TestInt_Equal(t *testing.T) {
	t.Run("", func(t *testing.T) {
		assert.True(t, NewInt(0).Equal(NewInt(0)))
		assert.False(t, NewInt(0).Equal(NewInt(1)))
		assert.True(t, NewInt(0).Equal(NewFloat32(0)))
	})
	t.Run("8", func(t *testing.T) {
		assert.True(t, NewInt8(0).Equal(NewInt(0)))
		assert.False(t, NewInt8(0).Equal(NewInt(1)))
		assert.True(t, NewInt8(0).Equal(NewFloat32(0)))
	})
	t.Run("16", func(t *testing.T) {
		assert.True(t, NewInt16(0).Equal(NewInt(0)))
		assert.False(t, NewInt16(0).Equal(NewInt(1)))
		assert.True(t, NewInt16(0).Equal(NewFloat32(0)))
	})
	t.Run("32", func(t *testing.T) {
		assert.True(t, NewInt32(0).Equal(NewInt(0)))
		assert.False(t, NewInt32(0).Equal(NewInt(1)))
		assert.True(t, NewInt32(0).Equal(NewFloat32(0)))
	})
	t.Run("64", func(t *testing.T) {
		assert.True(t, NewInt64(0).Equal(NewInt(0)))
		assert.False(t, NewInt64(0).Equal(NewInt(1)))
		assert.True(t, NewInt64(0).Equal(NewFloat32(0)))
	})
}

func TestInt_Compare(t *testing.T) {
	t.Run("", func(t *testing.T) {
		assert.Equal(t, 0, NewInt(0).Compare(NewInt(0)))
		assert.Equal(t, 1, NewInt(1).Compare(NewInt(0)))
		assert.Equal(t, -1, NewInt(0).Compare(NewInt(1)))
		assert.Equal(t, 0, NewInt(0).Compare(NewFloat32(0)))
	})
	t.Run("8", func(t *testing.T) {
		assert.Equal(t, 0, NewInt8(0).Compare(NewInt(0)))
		assert.Equal(t, 1, NewInt8(1).Compare(NewInt(0)))
		assert.Equal(t, -1, NewInt8(0).Compare(NewInt(1)))
		assert.Equal(t, 0, NewInt8(0).Compare(NewFloat32(0)))
	})
	t.Run("16", func(t *testing.T) {
		assert.Equal(t, 0, NewInt16(0).Compare(NewInt(0)))
		assert.Equal(t, 1, NewInt16(1).Compare(NewInt(0)))
		assert.Equal(t, -1, NewInt16(0).Compare(NewInt(1)))
		assert.Equal(t, 0, NewInt16(0).Compare(NewFloat32(0)))
	})
	t.Run("32", func(t *testing.T) {
		assert.Equal(t, 0, NewInt32(0).Compare(NewInt(0)))
		assert.Equal(t, 1, NewInt32(1).Compare(NewInt(0)))
		assert.Equal(t, -1, NewInt32(0).Compare(NewInt(1)))
		assert.Equal(t, 0, NewInt32(0).Compare(NewFloat32(0)))
	})
	t.Run("64", func(t *testing.T) {
		assert.Equal(t, 0, NewInt64(0).Compare(NewInt(0)))
		assert.Equal(t, 1, NewInt64(1).Compare(NewInt(0)))
		assert.Equal(t, -1, NewInt64(0).Compare(NewInt(1)))
		assert.Equal(t, 0, NewInt64(0).Compare(NewFloat32(0)))
	})
}

func TestInt_Hash(t *testing.T) {
	t.Run("", func(t *testing.T) {
		assert.NotEqual(t, NewInt(0).Hash(), NewInt(1).Hash())
		assert.Equal(t, NewInt(0).Hash(), NewInt(0).Hash())
		assert.Equal(t, NewInt(1).Hash(), NewInt(1).Hash())
	})
	t.Run("8", func(t *testing.T) {
		assert.NotEqual(t, NewInt8(0).Hash(), NewInt8(1).Hash())
		assert.Equal(t, NewInt8(0).Hash(), NewInt8(0).Hash())
		assert.Equal(t, NewInt8(1).Hash(), NewInt8(1).Hash())
	})
	t.Run("16", func(t *testing.T) {
		assert.NotEqual(t, NewInt16(0).Hash(), NewInt16(1).Hash())
		assert.Equal(t, NewInt16(0).Hash(), NewInt16(0).Hash())
		assert.Equal(t, NewInt16(1).Hash(), NewInt16(1).Hash())
	})
	t.Run("32", func(t *testing.T) {
		assert.NotEqual(t, NewInt32(0).Hash(), NewInt32(1).Hash())
		assert.Equal(t, NewInt32(0).Hash(), NewInt32(0).Hash())
		assert.Equal(t, NewInt32(1).Hash(), NewInt32(1).Hash())
	})
	t.Run("64", func(t *testing.T) {
		assert.NotEqual(t, NewInt64(0).Hash(), NewInt64(1).Hash())
		assert.Equal(t, NewInt64(0).Hash(), NewInt64(0).Hash())
		assert.Equal(t, NewInt64(1).Hash(), NewInt64(1).Hash())
	})
}

func TestInt_Encode(t *testing.T) {
	e := NewIntEncoder()

	t.Run("", func(t *testing.T) {
		v, err := e.Encode(int(1))
		assert.NoError(t, err)
		assert.Equal(t, NewInt(1), v)
	})
	t.Run("8", func(t *testing.T) {
		v, err := e.Encode(int8(1))
		assert.NoError(t, err)
		assert.Equal(t, NewInt8(1), v)
	})
	t.Run("16", func(t *testing.T) {
		v, err := e.Encode(int16(1))
		assert.NoError(t, err)
		assert.Equal(t, NewInt16(1), v)
	})
	t.Run("32", func(t *testing.T) {
		v, err := e.Encode(int32(1))
		assert.NoError(t, err)
		assert.Equal(t, NewInt32(1), v)
	})
	t.Run("64", func(t *testing.T) {
		v, err := e.Encode(int64(1))
		assert.NoError(t, err)
		assert.Equal(t, NewInt64(1), v)
	})
}

func TestInt_Decode(t *testing.T) {
	d := NewIntDecoder()

	t.Run("", func(t *testing.T) {
		var v int
		err := d.Decode(NewInt(1), &v)
		assert.NoError(t, err)
		assert.Equal(t, int(1), v)
	})
	t.Run("8", func(t *testing.T) {
		var v int8
		err := d.Decode(NewInt8(1), &v)
		assert.NoError(t, err)
		assert.Equal(t, int8(1), v)
	})
	t.Run("16", func(t *testing.T) {
		var v int16
		err := d.Decode(NewInt16(1), &v)
		assert.NoError(t, err)
		assert.Equal(t, int16(1), v)
	})
	t.Run("32", func(t *testing.T) {
		var v int32
		err := d.Decode(NewInt32(1), &v)
		assert.NoError(t, err)
		assert.Equal(t, int32(1), v)
	})
	t.Run("64", func(t *testing.T) {
		var v int64
		err := d.Decode(NewInt64(1), &v)
		assert.NoError(t, err)
		assert.Equal(t, int64(1), v)
	})
}
