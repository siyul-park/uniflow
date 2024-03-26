package primitive

import (
	"github.com/siyul-park/uniflow/pkg/encoding"
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

func TestFloat_Encode(t *testing.T) {
	enc := encoding.NewCompiledDecoder[*Value, any]()
	enc.Add(newFloatEncoder())

	t.Run("float32", func(t *testing.T) {
		source := float32(1)
		v := NewFloat32(source)

		var decoded Value
		err := enc.Decode(&decoded, &source)
		assert.NoError(t, err)
		assert.Equal(t, v, decoded)
	})

	t.Run("float64", func(t *testing.T) {
		source := float64(1)
		v := NewFloat64(source)

		var decoded Value
		err := enc.Decode(&decoded, &source)
		assert.NoError(t, err)
		assert.Equal(t, v, decoded)
	})
}

func TestFloat_Decode(t *testing.T) {
	dec := encoding.NewCompiledDecoder[Value, any]()
	dec.Add(newFloatDecoder())

	t.Run("float32", func(t *testing.T) {
		source := float32(1)
		v := NewFloat32(source)

		var decoded float32
		err := dec.Decode(v, &decoded)
		assert.NoError(t, err)
		assert.Equal(t, source, decoded)
	})

	t.Run("float64", func(t *testing.T) {
		source := float64(1)
		v := NewFloat64(source)

		var decoded float64
		err := dec.Decode(v, &decoded)
		assert.NoError(t, err)
		assert.Equal(t, source, decoded)
	})

	t.Run("int", func(t *testing.T) {
		source := float64(1)
		v := NewFloat64(source)

		var decoded int
		err := dec.Decode(v, &decoded)
		assert.NoError(t, err)
		assert.EqualValues(t, source, decoded)
	})

	t.Run("int8", func(t *testing.T) {
		source := float64(1)
		v := NewFloat64(source)

		var decoded int8
		err := dec.Decode(v, &decoded)
		assert.NoError(t, err)
		assert.EqualValues(t, source, decoded)
	})

	t.Run("int16", func(t *testing.T) {
		source := float64(1)
		v := NewFloat64(source)

		var decoded int16
		err := dec.Decode(v, &decoded)
		assert.NoError(t, err)
		assert.EqualValues(t, source, decoded)
	})

	t.Run("int32", func(t *testing.T) {
		source := float64(1)
		v := NewFloat64(source)

		var decoded int32
		err := dec.Decode(v, &decoded)
		assert.NoError(t, err)
		assert.EqualValues(t, source, decoded)
	})

	t.Run("int64", func(t *testing.T) {
		source := float64(1)
		v := NewFloat64(source)

		var decoded int64
		err := dec.Decode(v, &decoded)
		assert.NoError(t, err)
		assert.EqualValues(t, source, decoded)
	})

	t.Run("uint", func(t *testing.T) {
		source := float64(1)
		v := NewFloat64(source)

		var decoded uint
		err := dec.Decode(v, &decoded)
		assert.NoError(t, err)
		assert.EqualValues(t, source, decoded)
	})

	t.Run("uint8", func(t *testing.T) {
		source := float64(1)
		v := NewFloat64(source)

		var decoded uint8
		err := dec.Decode(v, &decoded)
		assert.NoError(t, err)
		assert.EqualValues(t, source, decoded)
	})

	t.Run("uint16", func(t *testing.T) {
		source := float64(1)
		v := NewFloat64(source)

		var decoded uint16
		err := dec.Decode(v, &decoded)
		assert.NoError(t, err)
		assert.EqualValues(t, source, decoded)
	})

	t.Run("uint32", func(t *testing.T) {
		source := float64(1)
		v := NewFloat64(source)

		var decoded uint32
		err := dec.Decode(v, &decoded)
		assert.NoError(t, err)
		assert.EqualValues(t, source, decoded)
	})

	t.Run("uint64", func(t *testing.T) {
		source := float64(1)
		v := NewFloat64(source)

		var decoded uint64
		err := dec.Decode(v, &decoded)
		assert.NoError(t, err)
		assert.EqualValues(t, source, decoded)
	})

	t.Run("any", func(t *testing.T) {
		source := float64(1)
		v := NewFloat64(source)

		var decoded any
		err := dec.Decode(v, &decoded)
		assert.NoError(t, err)
		assert.Equal(t, source, decoded)
	})
}
