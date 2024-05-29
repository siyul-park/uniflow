package object

import (
	"testing"

	"github.com/siyul-park/uniflow/pkg/encoding"

	"github.com/stretchr/testify/assert"
)

func TestUinteger(t *testing.T) {
	t.Run("Uint", func(t *testing.T) {
		v := NewUint(0)
		assert.Equal(t, KindUint, v.Kind())
		assert.Equal(t, uint(0), v.Interface())
		assert.Equal(t, uint64(0), v.Uint())
	})

	t.Run("Uint8", func(t *testing.T) {
		v := NewUint8(0)
		assert.Equal(t, KindUint8, v.Kind())
		assert.Equal(t, uint8(0), v.Interface())
		assert.Equal(t, uint64(0), v.Uint())
	})

	t.Run("Uint16", func(t *testing.T) {
		v := NewUint16(0)
		assert.Equal(t, KindUint16, v.Kind())
		assert.Equal(t, uint16(0), v.Interface())
		assert.Equal(t, uint64(0), v.Uint())
	})

	t.Run("Uint32", func(t *testing.T) {
		v := NewUint32(0)
		assert.Equal(t, KindUint32, v.Kind())
		assert.Equal(t, uint32(0), v.Interface())
		assert.Equal(t, uint64(0), v.Uint())
	})

	t.Run("Uint64", func(t *testing.T) {
		v := NewUint64(0)
		assert.Equal(t, KindUint64, v.Kind())
		assert.Equal(t, uint64(0), v.Interface())
		assert.Equal(t, uint64(0), v.Uint())
	})
}

func TestUinteger_Compare(t *testing.T) {
	t.Run("Uint", func(t *testing.T) {
		assert.Equal(t, 0, NewUint(0).Compare(NewUint(0)))
		assert.Equal(t, 1, NewUint(1).Compare(NewUint(0)))
		assert.Equal(t, -1, NewUint(0).Compare(NewUint(1)))
		assert.Equal(t, 0, NewUint(0).Compare(NewFloat32(0)))
	})

	t.Run("Uint8", func(t *testing.T) {
		assert.Equal(t, 0, NewUint8(0).Compare(NewUint(0)))
		assert.Equal(t, 1, NewUint8(1).Compare(NewUint(0)))
		assert.Equal(t, -1, NewUint8(0).Compare(NewUint(1)))
		assert.Equal(t, 0, NewUint8(0).Compare(NewFloat32(0)))
	})

	t.Run("Uint16", func(t *testing.T) {
		assert.Equal(t, 0, NewUint16(0).Compare(NewUint(0)))
		assert.Equal(t, 1, NewUint16(1).Compare(NewUint(0)))
		assert.Equal(t, -1, NewUint16(0).Compare(NewUint(1)))
		assert.Equal(t, 0, NewUint16(0).Compare(NewFloat32(0)))
	})

	t.Run("Uint32", func(t *testing.T) {
		assert.Equal(t, 0, NewUint32(0).Compare(NewUint(0)))
		assert.Equal(t, 1, NewUint32(1).Compare(NewUint(0)))
		assert.Equal(t, -1, NewUint32(0).Compare(NewUint(1)))
		assert.Equal(t, 0, NewUint32(0).Compare(NewFloat32(0)))
	})

	t.Run("Uint64", func(t *testing.T) {
		assert.Equal(t, 0, NewUint64(0).Compare(NewUint(0)))
		assert.Equal(t, 1, NewUint64(1).Compare(NewUint(0)))
		assert.Equal(t, -1, NewUint64(0).Compare(NewUint(1)))
		assert.Equal(t, 0, NewUint64(0).Compare(NewFloat32(0)))
	})
}

func TestUinteger_Encode(t *testing.T) {
	enc := encoding.NewAssembler[*Object, any]()
	enc.Add(newUintegerEncoder())

	t.Run("uint", func(t *testing.T) {
		source := uint(1)
		v := NewUint(source)

		var decoded Object
		err := enc.Encode(&decoded, &source)
		assert.NoError(t, err)
		assert.Equal(t, v, decoded)
	})

	t.Run("uint8", func(t *testing.T) {
		source := uint8(1)
		v := NewUint8(source)

		var decoded Object
		err := enc.Encode(&decoded, &source)
		assert.NoError(t, err)
		assert.Equal(t, v, decoded)
	})

	t.Run("uint16", func(t *testing.T) {
		source := uint16(1)
		v := NewUint16(source)

		var decoded Object
		err := enc.Encode(&decoded, &source)
		assert.NoError(t, err)
		assert.Equal(t, v, decoded)
	})

	t.Run("in32", func(t *testing.T) {
		source := uint32(1)
		v := NewUint32(source)

		var decoded Object
		err := enc.Encode(&decoded, &source)
		assert.NoError(t, err)
		assert.Equal(t, v, decoded)
	})

	t.Run("uint64", func(t *testing.T) {
		source := uint64(1)
		v := NewUint64(source)

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
		v := NewUint(source)

		var decoded float32
		err := dec.Encode(v, &decoded)
		assert.NoError(t, err)
		assert.EqualValues(t, source, decoded)
	})

	t.Run("float64", func(t *testing.T) {
		source := uint(1)
		v := NewUint(source)

		var decoded float64
		err := dec.Encode(v, &decoded)
		assert.NoError(t, err)
		assert.EqualValues(t, source, decoded)
	})

	t.Run("int", func(t *testing.T) {
		source := uint(1)
		v := NewUint(source)

		var decoded int
		err := dec.Encode(v, &decoded)
		assert.NoError(t, err)
		assert.EqualValues(t, source, decoded)
	})

	t.Run("int8", func(t *testing.T) {
		source := uint(1)
		v := NewUint(source)

		var decoded int8
		err := dec.Encode(v, &decoded)
		assert.NoError(t, err)
		assert.EqualValues(t, source, decoded)
	})

	t.Run("int16", func(t *testing.T) {
		source := uint(1)
		v := NewUint(source)

		var decoded int16
		err := dec.Encode(v, &decoded)
		assert.NoError(t, err)
		assert.EqualValues(t, source, decoded)
	})

	t.Run("int32", func(t *testing.T) {
		source := uint(1)
		v := NewUint(source)

		var decoded int32
		err := dec.Encode(v, &decoded)
		assert.NoError(t, err)
		assert.EqualValues(t, source, decoded)
	})

	t.Run("int64", func(t *testing.T) {
		source := uint(1)
		v := NewUint(source)

		var decoded int64
		err := dec.Encode(v, &decoded)
		assert.NoError(t, err)
		assert.EqualValues(t, source, decoded)
	})

	t.Run("uint", func(t *testing.T) {
		source := uint(1)
		v := NewUint(source)

		var decoded uint
		err := dec.Encode(v, &decoded)
		assert.NoError(t, err)
		assert.EqualValues(t, source, decoded)
	})

	t.Run("uint8", func(t *testing.T) {
		source := uint(1)
		v := NewUint(source)

		var decoded uint8
		err := dec.Encode(v, &decoded)
		assert.NoError(t, err)
		assert.EqualValues(t, source, decoded)
	})

	t.Run("uint16", func(t *testing.T) {
		source := uint(1)
		v := NewUint(source)

		var decoded uint16
		err := dec.Encode(v, &decoded)
		assert.NoError(t, err)
		assert.EqualValues(t, source, decoded)
	})

	t.Run("uint32", func(t *testing.T) {
		source := uint(1)
		v := NewUint(source)

		var decoded uint32
		err := dec.Encode(v, &decoded)
		assert.NoError(t, err)
		assert.EqualValues(t, source, decoded)
	})

	t.Run("uint64", func(t *testing.T) {
		source := uint(1)
		v := NewUint(source)

		var decoded uint64
		err := dec.Encode(v, &decoded)
		assert.NoError(t, err)
		assert.EqualValues(t, source, decoded)
	})

	t.Run("any", func(t *testing.T) {
		source := uint(1)
		v := NewUint(source)

		var decoded any
		err := dec.Encode(v, &decoded)
		assert.NoError(t, err)
		assert.Equal(t, source, decoded)
	})
}
