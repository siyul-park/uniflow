package object

import (
	"testing"

	"github.com/siyul-park/uniflow/pkg/encoding"

	"github.com/stretchr/testify/assert"
)

func TestNewInteger(t *testing.T) {
	t.Run("Int", func(t *testing.T) {
		v := NewInt(0)
		assert.Equal(t, KindInt, v.Kind())
		assert.Equal(t, int(0), v.Interface())
		assert.Equal(t, int64(0), v.Int())
	})

	t.Run("Int8", func(t *testing.T) {
		v := NewInt8(0)
		assert.Equal(t, KindInt8, v.Kind())
		assert.Equal(t, int8(0), v.Interface())
		assert.Equal(t, int64(0), v.Int())
	})

	t.Run("Int16", func(t *testing.T) {
		v := NewInt16(0)
		assert.Equal(t, KindInt16, v.Kind())
		assert.Equal(t, int16(0), v.Interface())
		assert.Equal(t, int64(0), v.Int())
	})

	t.Run("Int32", func(t *testing.T) {
		v := NewInt32(0)
		assert.Equal(t, KindInt32, v.Kind())
		assert.Equal(t, int32(0), v.Interface())
		assert.Equal(t, int64(0), v.Int())
	})

	t.Run("Int64", func(t *testing.T) {
		v := NewInt64(0)
		assert.Equal(t, KindInt64, v.Kind())
		assert.Equal(t, int64(0), v.Interface())
		assert.Equal(t, int64(0), v.Int())
	})
}

func TestInteger_Compare(t *testing.T) {
	t.Run("Int", func(t *testing.T) {
		assert.Equal(t, 0, NewInt(0).Compare(NewInt(0)))
		assert.Equal(t, 1, NewInt(1).Compare(NewInt(0)))
		assert.Equal(t, -1, NewInt(0).Compare(NewInt(1)))
		assert.Equal(t, 0, NewInt(0).Compare(NewFloat32(0)))
	})

	t.Run("Int8", func(t *testing.T) {
		assert.Equal(t, 0, NewInt8(0).Compare(NewInt(0)))
		assert.Equal(t, 1, NewInt8(1).Compare(NewInt(0)))
		assert.Equal(t, -1, NewInt8(0).Compare(NewInt(1)))
		assert.Equal(t, 0, NewInt8(0).Compare(NewFloat32(0)))
	})

	t.Run("Int16", func(t *testing.T) {
		assert.Equal(t, 0, NewInt16(0).Compare(NewInt(0)))
		assert.Equal(t, 1, NewInt16(1).Compare(NewInt(0)))
		assert.Equal(t, -1, NewInt16(0).Compare(NewInt(1)))
		assert.Equal(t, 0, NewInt16(0).Compare(NewFloat32(0)))
	})

	t.Run("Int32", func(t *testing.T) {
		assert.Equal(t, 0, NewInt32(0).Compare(NewInt(0)))
		assert.Equal(t, 1, NewInt32(1).Compare(NewInt(0)))
		assert.Equal(t, -1, NewInt32(0).Compare(NewInt(1)))
		assert.Equal(t, 0, NewInt32(0).Compare(NewFloat32(0)))
	})

	t.Run("Int64", func(t *testing.T) {
		assert.Equal(t, 0, NewInt64(0).Compare(NewInt(0)))
		assert.Equal(t, 1, NewInt64(1).Compare(NewInt(0)))
		assert.Equal(t, -1, NewInt64(0).Compare(NewInt(1)))
		assert.Equal(t, 0, NewInt64(0).Compare(NewFloat32(0)))
	})
}

func TestInteger_Encode(t *testing.T) {
	enc := encoding.NewAssembler[*Object, any]()
	enc.Add(newIntegerEncoder())

	t.Run("int", func(t *testing.T) {
		source := 1
		v := NewInt(source)

		var decoded Object
		err := enc.Encode(&decoded, &source)
		assert.NoError(t, err)
		assert.Equal(t, v, decoded)
	})

	t.Run("int8", func(t *testing.T) {
		source := int8(1)
		v := NewInt8(source)

		var decoded Object
		err := enc.Encode(&decoded, &source)
		assert.NoError(t, err)
		assert.Equal(t, v, decoded)
	})

	t.Run("int16", func(t *testing.T) {
		source := int16(1)
		v := NewInt16(source)

		var decoded Object
		err := enc.Encode(&decoded, &source)
		assert.NoError(t, err)
		assert.Equal(t, v, decoded)
	})

	t.Run("in32", func(t *testing.T) {
		source := int32(1)
		v := NewInt32(source)

		var decoded Object
		err := enc.Encode(&decoded, &source)
		assert.NoError(t, err)
		assert.Equal(t, v, decoded)
	})

	t.Run("int64", func(t *testing.T) {
		source := int64(1)
		v := NewInt64(source)

		var decoded Object
		err := enc.Encode(&decoded, &source)
		assert.NoError(t, err)
		assert.Equal(t, v, decoded)
	})
}

func TestInteger_Decode(t *testing.T) {
	dec := encoding.NewAssembler[Object, any]()
	dec.Add(newIntegerDecoder())

	t.Run("float32", func(t *testing.T) {
		source := 1
		v := NewInt(source)

		var decoded float32
		err := dec.Encode(v, &decoded)
		assert.NoError(t, err)
		assert.EqualValues(t, source, decoded)
	})

	t.Run("float64", func(t *testing.T) {
		source := 1
		v := NewInt(source)

		var decoded float64
		err := dec.Encode(v, &decoded)
		assert.NoError(t, err)
		assert.EqualValues(t, source, decoded)
	})

	t.Run("int", func(t *testing.T) {
		source := 1
		v := NewInt(source)

		var decoded int
		err := dec.Encode(v, &decoded)
		assert.NoError(t, err)
		assert.EqualValues(t, source, decoded)
	})

	t.Run("int8", func(t *testing.T) {
		source := 1
		v := NewInt(source)

		var decoded int8
		err := dec.Encode(v, &decoded)
		assert.NoError(t, err)
		assert.EqualValues(t, source, decoded)
	})

	t.Run("int16", func(t *testing.T) {
		source := 1
		v := NewInt(source)

		var decoded int16
		err := dec.Encode(v, &decoded)
		assert.NoError(t, err)
		assert.EqualValues(t, source, decoded)
	})

	t.Run("int32", func(t *testing.T) {
		source := 1
		v := NewInt(source)

		var decoded int32
		err := dec.Encode(v, &decoded)
		assert.NoError(t, err)
		assert.EqualValues(t, source, decoded)
	})

	t.Run("int64", func(t *testing.T) {
		source := 1
		v := NewInt(source)

		var decoded int64
		err := dec.Encode(v, &decoded)
		assert.NoError(t, err)
		assert.EqualValues(t, source, decoded)
	})

	t.Run("uint", func(t *testing.T) {
		source := 1
		v := NewInt(source)

		var decoded uint
		err := dec.Encode(v, &decoded)
		assert.NoError(t, err)
		assert.EqualValues(t, source, decoded)
	})

	t.Run("uint8", func(t *testing.T) {
		source := 1
		v := NewInt(source)

		var decoded uint8
		err := dec.Encode(v, &decoded)
		assert.NoError(t, err)
		assert.EqualValues(t, source, decoded)
	})

	t.Run("uint16", func(t *testing.T) {
		source := 1
		v := NewInt(source)

		var decoded uint16
		err := dec.Encode(v, &decoded)
		assert.NoError(t, err)
		assert.EqualValues(t, source, decoded)
	})

	t.Run("uint32", func(t *testing.T) {
		source := 1
		v := NewInt(source)

		var decoded uint32
		err := dec.Encode(v, &decoded)
		assert.NoError(t, err)
		assert.EqualValues(t, source, decoded)
	})

	t.Run("uint64", func(t *testing.T) {
		source := 1
		v := NewInt(source)

		var decoded uint64
		err := dec.Encode(v, &decoded)
		assert.NoError(t, err)
		assert.EqualValues(t, source, decoded)
	})

	t.Run("any", func(t *testing.T) {
		source := 1
		v := NewInt(source)

		var decoded any
		err := dec.Encode(v, &decoded)
		assert.NoError(t, err)
		assert.Equal(t, source, decoded)
	})
}
