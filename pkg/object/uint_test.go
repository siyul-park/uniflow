package object

import (
	"testing"

	"github.com/siyul-park/uniflow/pkg/encoding"

	"github.com/stretchr/testify/assert"
)

func TestUint_Uint(t *testing.T) {
	v := NewUint(42)

	assert.Equal(t, uint64(42), v.Uint())
}

func TestUint_Kind(t *testing.T) {
	v := NewUint(42)

	assert.Equal(t, KindUint, v.Kind())
}

func TestUint_Hash(t *testing.T) {
	v1 := NewUint(42)
	v2 := NewUint(24)

	assert.NotEqual(t, v1.Hash(), v2.Hash())
}

func TestUint_Interface(t *testing.T) {
	v := NewUint(42)

	assert.Equal(t, uint64(42), v.Interface())
}

func TestUint_Equal(t *testing.T) {
	v1 := NewUint(42)
	v2 := NewUint(24)

	assert.True(t, v1.Equal(v1))
	assert.True(t, v2.Equal(v2))
	assert.False(t, v1.Equal(v2))
}

func TestUint_Compare(t *testing.T) {
	v1 := NewUint(24)
	v2 := NewUint(42)

	assert.Equal(t, 0, v1.Compare(v1))
	assert.Equal(t, 0, v2.Compare(v2))
	assert.Equal(t, -1, v1.Compare(v2))
	assert.Equal(t, 1, v2.Compare(v1))
}

func TestUint_Encode(t *testing.T) {
	enc := encoding.NewEncodeAssembler[any, Object]()
	enc.Add(newUintEncoder())

	t.Run("uint", func(t *testing.T) {
		source := uint(1)
		v := NewUint(uint64(source))

		decoded, err := enc.Encode(&source)
		assert.NoError(t, err)
		assert.Equal(t, v, decoded)
	})

	t.Run("uint8", func(t *testing.T) {
		source := uint8(1)
		v := NewUint(uint64(source))

		decoded, err := enc.Encode(&source)
		assert.NoError(t, err)
		assert.Equal(t, v, decoded)
	})

	t.Run("uint16", func(t *testing.T) {
		source := uint16(1)
		v := NewUint(uint64(source))

		decoded, err := enc.Encode(&source)
		assert.NoError(t, err)
		assert.Equal(t, v, decoded)
	})

	t.Run("in32", func(t *testing.T) {
		source := uint32(1)
		v := NewUint(uint64(source))

		decoded, err := enc.Encode(&source)
		assert.NoError(t, err)
		assert.Equal(t, v, decoded)
	})

	t.Run("uint64", func(t *testing.T) {
		source := uint64(1)
		v := NewUint(source)

		decoded, err := enc.Encode(&source)
		assert.NoError(t, err)
		assert.Equal(t, v, decoded)
	})
}

func TestUint_Decode(t *testing.T) {
	dec := encoding.NewDecodeAssembler[Object, any]()
	dec.Add(newUintDecoder())

	t.Run("float32", func(t *testing.T) {
		source := uint(1)
		v := NewUint(uint64(source))

		var decoded float32
		err := dec.Decode(v, &decoded)
		assert.NoError(t, err)
		assert.EqualValues(t, source, decoded)
	})

	t.Run("float64", func(t *testing.T) {
		source := uint(1)
		v := NewUint(uint64(source))

		var decoded float64
		err := dec.Decode(v, &decoded)
		assert.NoError(t, err)
		assert.EqualValues(t, source, decoded)
	})

	t.Run("int", func(t *testing.T) {
		source := uint(1)
		v := NewUint(uint64(source))

		var decoded int
		err := dec.Decode(v, &decoded)
		assert.NoError(t, err)
		assert.EqualValues(t, source, decoded)
	})

	t.Run("int8", func(t *testing.T) {
		source := uint(1)
		v := NewUint(uint64(source))

		var decoded int8
		err := dec.Decode(v, &decoded)
		assert.NoError(t, err)
		assert.EqualValues(t, source, decoded)
	})

	t.Run("int16", func(t *testing.T) {
		source := uint(1)
		v := NewUint(uint64(source))

		var decoded int16
		err := dec.Decode(v, &decoded)
		assert.NoError(t, err)
		assert.EqualValues(t, source, decoded)
	})

	t.Run("int32", func(t *testing.T) {
		source := uint(1)
		v := NewUint(uint64(source))

		var decoded int32
		err := dec.Decode(v, &decoded)
		assert.NoError(t, err)
		assert.EqualValues(t, source, decoded)
	})

	t.Run("int64", func(t *testing.T) {
		source := uint(1)
		v := NewUint(uint64(source))

		var decoded int64
		err := dec.Decode(v, &decoded)
		assert.NoError(t, err)
		assert.EqualValues(t, source, decoded)
	})

	t.Run("uint", func(t *testing.T) {
		source := uint(1)
		v := NewUint(uint64(source))

		var decoded uint
		err := dec.Decode(v, &decoded)
		assert.NoError(t, err)
		assert.EqualValues(t, source, decoded)
	})

	t.Run("uint8", func(t *testing.T) {
		source := uint(1)
		v := NewUint(uint64(source))

		var decoded uint8
		err := dec.Decode(v, &decoded)
		assert.NoError(t, err)
		assert.EqualValues(t, source, decoded)
	})

	t.Run("uint16", func(t *testing.T) {
		source := uint(1)
		v := NewUint(uint64(source))

		var decoded uint16
		err := dec.Decode(v, &decoded)
		assert.NoError(t, err)
		assert.EqualValues(t, source, decoded)
	})

	t.Run("uint32", func(t *testing.T) {
		source := uint(1)
		v := NewUint(uint64(source))

		var decoded uint32
		err := dec.Decode(v, &decoded)
		assert.NoError(t, err)
		assert.EqualValues(t, source, decoded)
	})

	t.Run("uint64", func(t *testing.T) {
		source := uint(1)
		v := NewUint(uint64(source))

		var decoded uint64
		err := dec.Decode(v, &decoded)
		assert.NoError(t, err)
		assert.EqualValues(t, source, decoded)
	})
}
