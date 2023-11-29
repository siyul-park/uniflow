package primitive

import (
	"testing"

	"github.com/go-faker/faker/v4"
	"github.com/stretchr/testify/assert"
)

func TestNewSlice(t *testing.T) {
	v1 := NewString(faker.Word())

	o := NewSlice(v1)

	assert.Equal(t, KindSlice, o.Kind())
	assert.Equal(t, []string{v1.String()}, o.Interface())
}

func TestSlice_GetAndSet(t *testing.T) {
	v1 := NewString(faker.Word())
	v2 := NewString(faker.Word())

	o := NewSlice(v1)

	// Test Get
	r1 := o.Get(0)
	assert.Equal(t, v1, r1)

	// Test Get with out-of-bounds index
	r2 := o.Get(1)
	assert.Nil(t, r2)

	// Test Set
	o = o.Set(0, v2)

	// Test Get after Set
	r3 := o.Get(0)
	assert.Equal(t, v2, r3)
}

func TestSlice_Prepend(t *testing.T) {
	v := NewString(faker.Word())

	o := NewSlice()
	o = o.Prepend(v)

	assert.Equal(t, 1, o.Len())
}

func TestSlice_Append(t *testing.T) {
	v := NewString(faker.Word())

	o := NewSlice()
	o = o.Append(v)

	assert.Equal(t, 1, o.Len())
}

func TestSlice_Sub(t *testing.T) {
	v1 := NewString(faker.Word())
	v2 := NewString(faker.Word())

	o := NewSlice(v1, v2)
	o = o.Sub(0, 1)

	assert.Equal(t, 1, o.Len())
}

func TestSlice_Compare(t *testing.T) {
	v1 := NewString("1")
	v2 := NewString("2")

	// Test equal slices
	assert.Equal(t, 0, NewSlice(v1, v2).Compare(NewSlice(v1, v2)))

	// Test greater slice
	assert.Equal(t, 1, NewSlice(v2, v1).Compare(NewSlice(v1, v2)))

	// Test lesser slice
	assert.Equal(t, -1, NewSlice(v1, v2).Compare(NewSlice(v2, v1)))
}

func TestSlice_EncodeAndDecode(t *testing.T) {
	encoder := NewSliceEncoder(NewStringEncoder())
	decoder := NewSliceDecoder(NewStringDecoder())

	v1 := NewString(faker.Word())
	v2 := NewString(faker.Word())

	t.Run("Encode", func(t *testing.T) {
		// Test Encode
		encoded, err := encoder.Encode([]any{v1.Interface(), v2.Interface()})
		assert.NoError(t, err)
		assert.Equal(t, NewSlice(v1, v2), encoded)
	})

	t.Run("Decode", func(t *testing.T) {
		// Test Decode
		var decoded []any
		err := decoder.Decode(NewSlice(v1, v2), &decoded)
		assert.NoError(t, err)
		assert.Equal(t, []any{v1.Interface(), v2.Interface()}, decoded)
	})
}
