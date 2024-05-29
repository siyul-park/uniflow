package object

import (
	"testing"

	"github.com/siyul-park/uniflow/pkg/encoding"

	"github.com/stretchr/testify/assert"
)

func TestNewInteger(t *testing.T) {
	v := NewInteger(0)
	assert.Equal(t, KindInteger, v.Kind())
	assert.Equal(t, int64(0), v.Interface())
	assert.Equal(t, int64(0), v.Int())
}

func TestInteger_Compare(t *testing.T) {
	assert.Equal(t, 0, NewInteger(0).Compare(NewInteger(0)))
	assert.Equal(t, 1, NewInteger(1).Compare(NewInteger(0)))
	assert.Equal(t, -1, NewInteger(0).Compare(NewInteger(1)))
	assert.Equal(t, 0, NewInteger(0).Compare(NewFloat(0)))
}

func TestInteger_Encode(t *testing.T) {
	enc := encoding.NewAssembler[*Object, any]()
	enc.Add(NewIntegerEncoder())

	t.Run("int", func(t *testing.T) {
		source := 1
		v := NewInteger(int64(source))

		var decoded Object
		err := enc.Encode(&decoded, &source)
		assert.NoError(t, err)
		assert.Equal(t, v, decoded)
	})

	t.Run("int8", func(t *testing.T) {
		source := int8(1)
		v := NewInteger(int64(source))

		var decoded Object
		err := enc.Encode(&decoded, &source)
		assert.NoError(t, err)
		assert.Equal(t, v, decoded)
	})

	t.Run("int16", func(t *testing.T) {
		source := int16(1)
		v := NewInteger(int64(source))

		var decoded Object
		err := enc.Encode(&decoded, &source)
		assert.NoError(t, err)
		assert.Equal(t, v, decoded)
	})

	t.Run("in32", func(t *testing.T) {
		source := int32(1)
		v := NewInteger(int64(source))

		var decoded Object
		err := enc.Encode(&decoded, &source)
		assert.NoError(t, err)
		assert.Equal(t, v, decoded)
	})

	t.Run("int64", func(t *testing.T) {
		source := int64(1)
		v := NewInteger(source)

		var decoded Object
		err := enc.Encode(&decoded, &source)
		assert.NoError(t, err)
		assert.Equal(t, v, decoded)
	})
}

func TestInteger_Decode(t *testing.T) {
	dec := encoding.NewAssembler[Object, any]()
	dec.Add(NewIntegerDecoder())

	t.Run("float32", func(t *testing.T) {
		source := 1
		v := NewInteger(int64(source))

		var decoded float32
		err := dec.Encode(v, &decoded)
		assert.NoError(t, err)
		assert.EqualValues(t, source, decoded)
	})

	t.Run("float64", func(t *testing.T) {
		source := 1
		v := NewInteger(int64(source))

		var decoded float64
		err := dec.Encode(v, &decoded)
		assert.NoError(t, err)
		assert.EqualValues(t, source, decoded)
	})

	t.Run("int", func(t *testing.T) {
		source := 1
		v := NewInteger(int64(source))

		var decoded int
		err := dec.Encode(v, &decoded)
		assert.NoError(t, err)
		assert.EqualValues(t, source, decoded)
	})

	t.Run("int8", func(t *testing.T) {
		source := 1
		v := NewInteger(int64(source))

		var decoded int8
		err := dec.Encode(v, &decoded)
		assert.NoError(t, err)
		assert.EqualValues(t, source, decoded)
	})

	t.Run("int16", func(t *testing.T) {
		source := 1
		v := NewInteger(int64(source))

		var decoded int16
		err := dec.Encode(v, &decoded)
		assert.NoError(t, err)
		assert.EqualValues(t, source, decoded)
	})

	t.Run("int32", func(t *testing.T) {
		source := 1
		v := NewInteger(int64(source))

		var decoded int32
		err := dec.Encode(v, &decoded)
		assert.NoError(t, err)
		assert.EqualValues(t, source, decoded)
	})

	t.Run("int64", func(t *testing.T) {
		source := 1
		v := NewInteger(int64(source))

		var decoded int64
		err := dec.Encode(v, &decoded)
		assert.NoError(t, err)
		assert.EqualValues(t, source, decoded)
	})

	t.Run("uint", func(t *testing.T) {
		source := 1
		v := NewInteger(int64(source))

		var decoded uint
		err := dec.Encode(v, &decoded)
		assert.NoError(t, err)
		assert.EqualValues(t, source, decoded)
	})

	t.Run("uint8", func(t *testing.T) {
		source := 1
		v := NewInteger(int64(source))

		var decoded uint8
		err := dec.Encode(v, &decoded)
		assert.NoError(t, err)
		assert.EqualValues(t, source, decoded)
	})

	t.Run("uint16", func(t *testing.T) {
		source := 1
		v := NewInteger(int64(source))

		var decoded uint16
		err := dec.Encode(v, &decoded)
		assert.NoError(t, err)
		assert.EqualValues(t, source, decoded)
	})

	t.Run("uint32", func(t *testing.T) {
		source := 1
		v := NewInteger(int64(source))

		var decoded uint32
		err := dec.Encode(v, &decoded)
		assert.NoError(t, err)
		assert.EqualValues(t, source, decoded)
	})

	t.Run("uint64", func(t *testing.T) {
		source := 1
		v := NewInteger(int64(source))

		var decoded uint64
		err := dec.Encode(v, &decoded)
		assert.NoError(t, err)
		assert.EqualValues(t, source, decoded)
	})
}
