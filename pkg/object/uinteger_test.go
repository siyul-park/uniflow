package object

import (
	"testing"

	"github.com/siyul-park/uniflow/pkg/encoding"

	"github.com/stretchr/testify/assert"
)

func TestUinteger(t *testing.T) {
	v := NewUInteger(1)
	assert.Equal(t, KindUInteger, v.Kind())
	assert.NotEqual(t, uint64(0), v.Hash())
	assert.Equal(t, uint64(1), v.Interface())
	assert.Equal(t, uint64(1), v.Uint())
}

func TestUinteger_Compare(t *testing.T) {
	assert.Equal(t, 0, NewUInteger(0).Compare(NewUInteger(0)))
	assert.Equal(t, 1, NewUInteger(1).Compare(NewUInteger(0)))
	assert.Equal(t, -1, NewUInteger(0).Compare(NewUInteger(1)))
	assert.Equal(t, 0, NewUInteger(0).Compare(NewFloat(0)))
}

func TestUinteger_Encode(t *testing.T) {
	enc := encoding.NewAssembler[*Object, any]()
	enc.Add(newUIntegerEncoder())

	t.Run("uint", func(t *testing.T) {
		source := uint(1)
		v := NewUInteger(uint64(source))

		var decoded Object
		err := enc.Encode(&decoded, &source)
		assert.NoError(t, err)
		assert.Equal(t, v, decoded)
	})

	t.Run("uint8", func(t *testing.T) {
		source := uint8(1)
		v := NewUInteger(uint64(source))

		var decoded Object
		err := enc.Encode(&decoded, &source)
		assert.NoError(t, err)
		assert.Equal(t, v, decoded)
	})

	t.Run("uint16", func(t *testing.T) {
		source := uint16(1)
		v := NewUInteger(uint64(source))

		var decoded Object
		err := enc.Encode(&decoded, &source)
		assert.NoError(t, err)
		assert.Equal(t, v, decoded)
	})

	t.Run("in32", func(t *testing.T) {
		source := uint32(1)
		v := NewUInteger(uint64(source))

		var decoded Object
		err := enc.Encode(&decoded, &source)
		assert.NoError(t, err)
		assert.Equal(t, v, decoded)
	})

	t.Run("uint64", func(t *testing.T) {
		source := uint64(1)
		v := NewUInteger(source)

		var decoded Object
		err := enc.Encode(&decoded, &source)
		assert.NoError(t, err)
		assert.Equal(t, v, decoded)
	})
}

func TestUinteger_Decode(t *testing.T) {
	dec := encoding.NewAssembler[Object, any]()
	dec.Add(newUintegerDecoder())

	t.Run("float32", func(t *testing.T) {
		source := uint(1)
		v := NewUInteger(uint64(source))

		var decoded float32
		err := dec.Encode(v, &decoded)
		assert.NoError(t, err)
		assert.EqualValues(t, source, decoded)
	})

	t.Run("float64", func(t *testing.T) {
		source := uint(1)
		v := NewUInteger(uint64(source))

		var decoded float64
		err := dec.Encode(v, &decoded)
		assert.NoError(t, err)
		assert.EqualValues(t, source, decoded)
	})

	t.Run("int", func(t *testing.T) {
		source := uint(1)
		v := NewUInteger(uint64(source))

		var decoded int
		err := dec.Encode(v, &decoded)
		assert.NoError(t, err)
		assert.EqualValues(t, source, decoded)
	})

	t.Run("int8", func(t *testing.T) {
		source := uint(1)
		v := NewUInteger(uint64(source))

		var decoded int8
		err := dec.Encode(v, &decoded)
		assert.NoError(t, err)
		assert.EqualValues(t, source, decoded)
	})

	t.Run("int16", func(t *testing.T) {
		source := uint(1)
		v := NewUInteger(uint64(source))

		var decoded int16
		err := dec.Encode(v, &decoded)
		assert.NoError(t, err)
		assert.EqualValues(t, source, decoded)
	})

	t.Run("int32", func(t *testing.T) {
		source := uint(1)
		v := NewUInteger(uint64(source))

		var decoded int32
		err := dec.Encode(v, &decoded)
		assert.NoError(t, err)
		assert.EqualValues(t, source, decoded)
	})

	t.Run("int64", func(t *testing.T) {
		source := uint(1)
		v := NewUInteger(uint64(source))

		var decoded int64
		err := dec.Encode(v, &decoded)
		assert.NoError(t, err)
		assert.EqualValues(t, source, decoded)
	})

	t.Run("uint", func(t *testing.T) {
		source := uint(1)
		v := NewUInteger(uint64(source))

		var decoded uint
		err := dec.Encode(v, &decoded)
		assert.NoError(t, err)
		assert.EqualValues(t, source, decoded)
	})

	t.Run("uint8", func(t *testing.T) {
		source := uint(1)
		v := NewUInteger(uint64(source))

		var decoded uint8
		err := dec.Encode(v, &decoded)
		assert.NoError(t, err)
		assert.EqualValues(t, source, decoded)
	})

	t.Run("uint16", func(t *testing.T) {
		source := uint(1)
		v := NewUInteger(uint64(source))

		var decoded uint16
		err := dec.Encode(v, &decoded)
		assert.NoError(t, err)
		assert.EqualValues(t, source, decoded)
	})

	t.Run("uint32", func(t *testing.T) {
		source := uint(1)
		v := NewUInteger(uint64(source))

		var decoded uint32
		err := dec.Encode(v, &decoded)
		assert.NoError(t, err)
		assert.EqualValues(t, source, decoded)
	})

	t.Run("uint64", func(t *testing.T) {
		source := uint(1)
		v := NewUInteger(uint64(source))

		var decoded uint64
		err := dec.Encode(v, &decoded)
		assert.NoError(t, err)
		assert.EqualValues(t, source, decoded)
	})
}
