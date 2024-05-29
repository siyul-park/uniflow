package object

import (
	"testing"

	"github.com/siyul-park/uniflow/pkg/encoding"

	"github.com/stretchr/testify/assert"
)

func TestNewFloat(t *testing.T) {
	v := NewFloat(0)

	assert.Equal(t, KindFloat, v.Kind())
	assert.Equal(t, float64(0), v.Interface())
	assert.Equal(t, float64(0), v.Float())
}

func TestFloat_Compare(t *testing.T) {
	assert.Equal(t, 0, NewFloat(0).Compare(NewFloat(0)))
	assert.Equal(t, 1, NewFloat(1).Compare(NewFloat(0)))
	assert.Equal(t, -1, NewFloat(0).Compare(NewFloat(1)))
}

func TestFloat_Encode(t *testing.T) {
	enc := encoding.NewAssembler[*Object, any]()
	enc.Add(newFloatEncoder())

	t.Run("float32", func(t *testing.T) {
		source := float32(1)
		v := NewFloat(float64(source))

		var decoded Object
		err := enc.Encode(&decoded, &source)
		assert.NoError(t, err)
		assert.Equal(t, v, decoded)
	})

	t.Run("float64", func(t *testing.T) {
		source := float64(1)
		v := NewFloat(source)

		var decoded Object
		err := enc.Encode(&decoded, &source)
		assert.NoError(t, err)
		assert.Equal(t, v, decoded)
	})
}

func TestFloat_Decode(t *testing.T) {
	dec := encoding.NewAssembler[Object, any]()
	dec.Add(newFloatDecoder())

	t.Run("float32", func(t *testing.T) {
		source := float32(1)
		v := NewFloat(float64(source))

		var decoded float32
		err := dec.Encode(v, &decoded)
		assert.NoError(t, err)
		assert.Equal(t, source, decoded)
	})

	t.Run("float64", func(t *testing.T) {
		source := float64(1)
		v := NewFloat(source)

		var decoded float64
		err := dec.Encode(v, &decoded)
		assert.NoError(t, err)
		assert.Equal(t, source, decoded)
	})

	t.Run("int", func(t *testing.T) {
		source := float64(1)
		v := NewFloat(source)

		var decoded int
		err := dec.Encode(v, &decoded)
		assert.NoError(t, err)
		assert.EqualValues(t, source, decoded)
	})

	t.Run("int8", func(t *testing.T) {
		source := float64(1)
		v := NewFloat(source)

		var decoded int8
		err := dec.Encode(v, &decoded)
		assert.NoError(t, err)
		assert.EqualValues(t, source, decoded)
	})

	t.Run("int16", func(t *testing.T) {
		source := float64(1)
		v := NewFloat(source)

		var decoded int16
		err := dec.Encode(v, &decoded)
		assert.NoError(t, err)
		assert.EqualValues(t, source, decoded)
	})

	t.Run("int32", func(t *testing.T) {
		source := float64(1)
		v := NewFloat(source)

		var decoded int32
		err := dec.Encode(v, &decoded)
		assert.NoError(t, err)
		assert.EqualValues(t, source, decoded)
	})

	t.Run("int64", func(t *testing.T) {
		source := float64(1)
		v := NewFloat(source)

		var decoded int64
		err := dec.Encode(v, &decoded)
		assert.NoError(t, err)
		assert.EqualValues(t, source, decoded)
	})

	t.Run("uint", func(t *testing.T) {
		source := float64(1)
		v := NewFloat(source)

		var decoded uint
		err := dec.Encode(v, &decoded)
		assert.NoError(t, err)
		assert.EqualValues(t, source, decoded)
	})

	t.Run("uint8", func(t *testing.T) {
		source := float64(1)
		v := NewFloat(source)

		var decoded uint8
		err := dec.Encode(v, &decoded)
		assert.NoError(t, err)
		assert.EqualValues(t, source, decoded)
	})

	t.Run("uint16", func(t *testing.T) {
		source := float64(1)
		v := NewFloat(source)

		var decoded uint16
		err := dec.Encode(v, &decoded)
		assert.NoError(t, err)
		assert.EqualValues(t, source, decoded)
	})

	t.Run("uint32", func(t *testing.T) {
		source := float64(1)
		v := NewFloat(source)

		var decoded uint32
		err := dec.Encode(v, &decoded)
		assert.NoError(t, err)
		assert.EqualValues(t, source, decoded)
	})

	t.Run("uint64", func(t *testing.T) {
		source := float64(1)
		v := NewFloat(source)

		var decoded uint64
		err := dec.Encode(v, &decoded)
		assert.NoError(t, err)
		assert.EqualValues(t, source, decoded)
	})

	t.Run("any", func(t *testing.T) {
		source := float64(1)
		v := NewFloat(source)

		var decoded any
		err := dec.Encode(v, &decoded)
		assert.NoError(t, err)
		assert.Equal(t, source, decoded)
	})
}
