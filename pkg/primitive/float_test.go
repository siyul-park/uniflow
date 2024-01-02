package primitive

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewFloat(t *testing.T) {
	t.Run("Float32", func(t *testing.T) {
		v := NewFloat32(0)

		assert.Equal(t, KindFloat32, v.Kind())
		assert.Equal(t, float32(0), v.Interface())
		assert.Equal(t, float64(0), v.Float())
	})

	t.Run("Float64", func(t *testing.T) {
		v := NewFloat64(0)

		assert.Equal(t, KindFloat64, v.Kind())
		assert.Equal(t, float64(0), v.Interface())
		assert.Equal(t, float64(0), v.Float())
	})
}

func TestFloat_Compare(t *testing.T) {
	t.Run("Float32", func(t *testing.T) {
		assert.Equal(t, 0, NewFloat32(0).Compare(NewFloat32(0)))
		assert.Equal(t, 0, NewFloat32(0).Compare(NewFloat64(0)))
		assert.Equal(t, 1, NewFloat32(1).Compare(NewFloat32(0)))
		assert.Equal(t, -1, NewFloat32(0).Compare(NewFloat32(1)))
	})

	t.Run("Float64", func(t *testing.T) {
		assert.Equal(t, 0, NewFloat64(0).Compare(NewFloat64(0)))
		assert.Equal(t, 0, NewFloat64(0).Compare(NewFloat32(0)))
		assert.Equal(t, 1, NewFloat64(1).Compare(NewFloat64(0)))
		assert.Equal(t, -1, NewFloat64(0).Compare(NewFloat64(1)))
	})
}

func TestFloat_EncodeAndDecode(t *testing.T) {
	e := NewFloatEncoder()
	d := NewFloatDecoder()

	t.Run("Float32", func(t *testing.T) {
		source := float32(1)

		encoded, err := e.Encode(source)
		assert.NoError(t, err)
		assert.Equal(t, Float32(1), encoded)

		var decoded float32
		err = d.Decode(encoded, &decoded)
		assert.NoError(t, err)
		assert.Equal(t, source, decoded)
	})

	t.Run("Float64", func(t *testing.T) {
		source := float64(1)

		encoded, err := e.Encode(source)
		assert.NoError(t, err)
		assert.Equal(t, Float64(1), encoded)

		var decoded float64
		err = d.Decode(encoded, &decoded)
		assert.NoError(t, err)
		assert.Equal(t, source, decoded)
	})
}
