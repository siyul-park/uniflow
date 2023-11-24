package primitive

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewFloat(t *testing.T) {
	t.Run("32", func(t *testing.T) {
		v := NewFloat32(0)

		assert.Equal(t, KindFloat32, v.Kind())
		assert.Equal(t, float32(0), v.Interface())
	})
	t.Run("64", func(t *testing.T) {
		v := NewFloat64(0)

		assert.Equal(t, KindFloat64, v.Kind())
		assert.Equal(t, float64(0), v.Interface())
	})
}

func TestFloat_Equal(t *testing.T) {
	t.Run("32", func(t *testing.T) {
		assert.True(t, NewFloat32(0).Equal(NewFloat32(0)))
		assert.True(t, NewFloat32(0).Equal(NewFloat64(0)))
		assert.False(t, NewFloat32(0).Equal(NewFloat32(1)))
	})

	t.Run("64", func(t *testing.T) {
		assert.True(t, NewFloat64(0).Equal(NewFloat64(0)))
		assert.True(t, NewFloat64(0).Equal(NewFloat32(0)))
		assert.False(t, NewFloat64(1).Equal(NewFloat64(0)))
	})
}

func TestFloat_Hash(t *testing.T) {
	t.Run("32", func(t *testing.T) {
		assert.NotEqual(t, NewFloat32(0).Hash(), NewFloat32(1).Hash())
		assert.Equal(t, NewFloat32(0).Hash(), NewFloat32(0).Hash())
		assert.Equal(t, NewFloat32(1).Hash(), NewFloat32(1).Hash())
	})

	t.Run("64", func(t *testing.T) {
		assert.NotEqual(t, NewFloat64(0).Hash(), NewFloat64(1).Hash())
		assert.Equal(t, NewFloat64(0).Hash(), NewFloat64(0).Hash())
		assert.Equal(t, NewFloat64(1).Hash(), NewFloat64(1).Hash())
	})
}

func TestFloat_Encode(t *testing.T) {
	e := NewFloatEncoder()

	t.Run("32", func(t *testing.T) {
		v, err := e.Encode(float32(1))
		assert.NoError(t, err)
		assert.Equal(t, NewFloat32(1), v)
	})
	t.Run("64", func(t *testing.T) {
		v, err := e.Encode(float64(1))
		assert.NoError(t, err)
		assert.Equal(t, NewFloat64(1), v)
	})
}

func TestFloat_Decode(t *testing.T) {
	d := NewFloatDecoder()

	t.Run("32", func(t *testing.T) {
		var v float32
		err := d.Decode(NewFloat32(1), &v)
		assert.NoError(t, err)
		assert.Equal(t, float32(1), v)
	})
	t.Run("64", func(t *testing.T) {
		var v float64
		err := d.Decode(NewFloat64(1), &v)
		assert.NoError(t, err)
		assert.Equal(t, float64(1), v)
	})
}
