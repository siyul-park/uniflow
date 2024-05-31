package object

import (
	"testing"

	"github.com/siyul-park/uniflow/pkg/encoding"

	"github.com/stretchr/testify/assert"
)

func TestInt_Int(t *testing.T) {
	v := NewInt(42)

	assert.Equal(t, int64(42), v.Int())
}

func TestInt_Kind(t *testing.T) {
	v := NewInt(42)

	assert.Equal(t, KindInt, v.Kind())
}

func TestInt_Hash(t *testing.T) {
	v1 := NewInt(42)
	v2 := NewInt(24)

	assert.NotEqual(t, v1.Hash(), v2.Hash())
}

func TestInt_Interface(t *testing.T) {
	v := NewInt(42)

	assert.Equal(t, int64(42), v.Interface())
}

func TestInt_Equal(t *testing.T) {
	v1 := NewInt(42)
	v2 := NewInt(24)

	assert.True(t, v1.Equal(v1))
	assert.True(t, v2.Equal(v2))
	assert.False(t, v1.Equal(v2))
}

func TestInt_Compare(t *testing.T) {
	v1 := NewInt(24)
	v2 := NewInt(42)

	assert.Equal(t, 0, v1.Compare(v1))
	assert.Equal(t, 0, v2.Compare(v2))
	assert.Equal(t, -1, v1.Compare(v2))
	assert.Equal(t, 1, v2.Compare(v1))
}

func TestInt_Encode(t *testing.T) {
	enc := encoding.NewEncodeAssembler[any, Object]()
	enc.Add(NewIntEncoder())

	t.Run("int", func(t *testing.T) {
		source := 1
		v := NewInt(int64(source))

		decoded, err := enc.Encode(&source)
		assert.NoError(t, err)
		assert.Equal(t, v, decoded)
	})

	t.Run("int8", func(t *testing.T) {
		source := int8(1)
		v := NewInt(int64(source))

		decoded, err := enc.Encode(&source)
		assert.NoError(t, err)
		assert.Equal(t, v, decoded)
	})

	t.Run("int16", func(t *testing.T) {
		source := int16(1)
		v := NewInt(int64(source))

		decoded, err := enc.Encode(&source)
		assert.NoError(t, err)
		assert.Equal(t, v, decoded)
	})

	t.Run("in32", func(t *testing.T) {
		source := int32(1)
		v := NewInt(int64(source))

		decoded, err := enc.Encode(&source)
		assert.NoError(t, err)
		assert.Equal(t, v, decoded)
	})

	t.Run("int64", func(t *testing.T) {
		source := int64(1)
		v := NewInt(source)

		decoded, err := enc.Encode(&source)
		assert.NoError(t, err)
		assert.Equal(t, v, decoded)
	})
}

func TestInt_Decode(t *testing.T) {
	dec := encoding.NewDecodeAssembler[Object, any]()
	dec.Add(NewIntDecoder())

	t.Run("float32", func(t *testing.T) {
		source := 1
		v := NewInt(int64(source))

		var decoded float32
		err := dec.Decode(v, &decoded)
		assert.NoError(t, err)
		assert.EqualValues(t, source, decoded)
	})

	t.Run("float64", func(t *testing.T) {
		source := 1
		v := NewInt(int64(source))

		var decoded float64
		err := dec.Decode(v, &decoded)
		assert.NoError(t, err)
		assert.EqualValues(t, source, decoded)
	})

	t.Run("int", func(t *testing.T) {
		source := 1
		v := NewInt(int64(source))

		var decoded int
		err := dec.Decode(v, &decoded)
		assert.NoError(t, err)
		assert.EqualValues(t, source, decoded)
	})

	t.Run("int8", func(t *testing.T) {
		source := 1
		v := NewInt(int64(source))

		var decoded int8
		err := dec.Decode(v, &decoded)
		assert.NoError(t, err)
		assert.EqualValues(t, source, decoded)
	})

	t.Run("int16", func(t *testing.T) {
		source := 1
		v := NewInt(int64(source))

		var decoded int16
		err := dec.Decode(v, &decoded)
		assert.NoError(t, err)
		assert.EqualValues(t, source, decoded)
	})

	t.Run("int32", func(t *testing.T) {
		source := 1
		v := NewInt(int64(source))

		var decoded int32
		err := dec.Decode(v, &decoded)
		assert.NoError(t, err)
		assert.EqualValues(t, source, decoded)
	})

	t.Run("int64", func(t *testing.T) {
		source := 1
		v := NewInt(int64(source))

		var decoded int64
		err := dec.Decode(v, &decoded)
		assert.NoError(t, err)
		assert.EqualValues(t, source, decoded)
	})

	t.Run("uint", func(t *testing.T) {
		source := 1
		v := NewInt(int64(source))

		var decoded uint
		err := dec.Decode(v, &decoded)
		assert.NoError(t, err)
		assert.EqualValues(t, source, decoded)
	})

	t.Run("uint8", func(t *testing.T) {
		source := 1
		v := NewInt(int64(source))

		var decoded uint8
		err := dec.Decode(v, &decoded)
		assert.NoError(t, err)
		assert.EqualValues(t, source, decoded)
	})

	t.Run("uint16", func(t *testing.T) {
		source := 1
		v := NewInt(int64(source))

		var decoded uint16
		err := dec.Decode(v, &decoded)
		assert.NoError(t, err)
		assert.EqualValues(t, source, decoded)
	})

	t.Run("uint32", func(t *testing.T) {
		source := 1
		v := NewInt(int64(source))

		var decoded uint32
		err := dec.Decode(v, &decoded)
		assert.NoError(t, err)
		assert.EqualValues(t, source, decoded)
	})

	t.Run("uint64", func(t *testing.T) {
		source := 1
		v := NewInt(int64(source))

		var decoded uint64
		err := dec.Decode(v, &decoded)
		assert.NoError(t, err)
		assert.EqualValues(t, source, decoded)
	})
}
