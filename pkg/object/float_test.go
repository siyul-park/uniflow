package object

import (
	"testing"

	"github.com/siyul-park/uniflow/pkg/encoding"

	"github.com/stretchr/testify/assert"
)

func TestFloat_Float(t *testing.T) {
	v := NewFloat(3.14)

	assert.Equal(t, 3.14, v.Float())
}

func TestFloat_Kind(t *testing.T) {
	v := NewFloat(3.14)

	assert.Equal(t, KindFloat, v.Kind())
}

func TestFloat_Hash(t *testing.T) {
	v1 := NewFloat(3.14)
	v2 := NewFloat(6.28)

	assert.NotEqual(t, v1.Hash(), v2.Hash())
}

func TestFloat_Interface(t *testing.T) {
	v := NewFloat(3.14)

	assert.Equal(t, 3.14, v.Interface())
}

func TestFloat_Equal(t *testing.T) {
	v1 := NewFloat(3.14)
	v2 := NewFloat(6.28)

	assert.True(t, v1.Equal(v1))
	assert.True(t, v2.Equal(v2))
	assert.False(t, v1.Equal(v2))
}

func TestFloat_Compare(t *testing.T) {
	v1 := NewFloat(3.14)
	v2 := NewFloat(6.28)

	assert.Equal(t, 0, v1.Compare(v1))
	assert.Equal(t, 0, v2.Compare(v2))
	assert.Equal(t, -1, v1.Compare(v2))
	assert.Equal(t, 1, v2.Compare(v1))
}

func TestFloat_Encode(t *testing.T) {
	enc := encoding.NewEncodeAssembler[any, Object]()
	enc.Add(newFloatEncoder())

	t.Run("float32", func(t *testing.T) {
		source := float32(1)
		v := NewFloat(float64(source))

		decoded, err := enc.Encode(source)
		assert.NoError(t, err)
		assert.Equal(t, v, decoded)
	})

	t.Run("float64", func(t *testing.T) {
		source := float64(1)
		v := NewFloat(source)

		decoded, err := enc.Encode(source)
		assert.NoError(t, err)
		assert.Equal(t, v, decoded)
	})
}

func TestFloat_Decode(t *testing.T) {
	dec := encoding.NewDecodeAssembler[Object, any]()
	dec.Add(newFloatDecoder())

	t.Run("float32", func(t *testing.T) {
		source := float32(1)
		v := NewFloat(float64(source))

		var decoded float32
		err := dec.Decode(v, &decoded)
		assert.NoError(t, err)
		assert.Equal(t, source, decoded)
	})

	t.Run("float64", func(t *testing.T) {
		source := float64(1)
		v := NewFloat(source)

		var decoded float64
		err := dec.Decode(v, &decoded)
		assert.NoError(t, err)
		assert.Equal(t, source, decoded)
	})

	t.Run("int", func(t *testing.T) {
		source := float64(1)
		v := NewFloat(source)

		var decoded int
		err := dec.Decode(v, &decoded)
		assert.NoError(t, err)
		assert.EqualValues(t, source, decoded)
	})

	t.Run("int8", func(t *testing.T) {
		source := float64(1)
		v := NewFloat(source)

		var decoded int8
		err := dec.Decode(v, &decoded)
		assert.NoError(t, err)
		assert.EqualValues(t, source, decoded)
	})

	t.Run("int16", func(t *testing.T) {
		source := float64(1)
		v := NewFloat(source)

		var decoded int16
		err := dec.Decode(v, &decoded)
		assert.NoError(t, err)
		assert.EqualValues(t, source, decoded)
	})

	t.Run("int32", func(t *testing.T) {
		source := float64(1)
		v := NewFloat(source)

		var decoded int32
		err := dec.Decode(v, &decoded)
		assert.NoError(t, err)
		assert.EqualValues(t, source, decoded)
	})

	t.Run("int64", func(t *testing.T) {
		source := float64(1)
		v := NewFloat(source)

		var decoded int64
		err := dec.Decode(v, &decoded)
		assert.NoError(t, err)
		assert.EqualValues(t, source, decoded)
	})

	t.Run("uint", func(t *testing.T) {
		source := float64(1)
		v := NewFloat(source)

		var decoded uint
		err := dec.Decode(v, &decoded)
		assert.NoError(t, err)
		assert.EqualValues(t, source, decoded)
	})

	t.Run("uint8", func(t *testing.T) {
		source := float64(1)
		v := NewFloat(source)

		var decoded uint8
		err := dec.Decode(v, &decoded)
		assert.NoError(t, err)
		assert.EqualValues(t, source, decoded)
	})

	t.Run("uint16", func(t *testing.T) {
		source := float64(1)
		v := NewFloat(source)

		var decoded uint16
		err := dec.Decode(v, &decoded)
		assert.NoError(t, err)
		assert.EqualValues(t, source, decoded)
	})

	t.Run("uint32", func(t *testing.T) {
		source := float64(1)
		v := NewFloat(source)

		var decoded uint32
		err := dec.Decode(v, &decoded)
		assert.NoError(t, err)
		assert.EqualValues(t, source, decoded)
	})

	t.Run("uint64", func(t *testing.T) {
		source := float64(1)
		v := NewFloat(source)

		var decoded uint64
		err := dec.Decode(v, &decoded)
		assert.NoError(t, err)
		assert.EqualValues(t, source, decoded)
	})

	t.Run("any", func(t *testing.T) {
		source := float64(1)
		v := NewFloat(source)

		var decoded any
		err := dec.Decode(v, &decoded)
		assert.NoError(t, err)
		assert.Equal(t, source, decoded)
	})
}
