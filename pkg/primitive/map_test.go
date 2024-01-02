package primitive

import (
	"testing"

	"github.com/go-faker/faker/v4"
	"github.com/stretchr/testify/assert"
)

func TestNewMap(t *testing.T) {
	k1 := NewString(faker.Word())
	v1 := NewString(faker.Word())

	o := NewMap(k1, v1)

	assert.Equal(t, KindMap, o.Kind())
	assert.Equal(t, map[string]string{k1.String(): v1.String()}, o.Interface())
	assert.Equal(t, map[any]any{k1.Interface(): v1.Interface()}, o.Map())
}

func TestMap_GetAndSetAndDelete(t *testing.T) {
	k1 := NewString(faker.Word())
	v1 := NewString(faker.Word())

	o := NewMap()
	o = o.Set(k1, v1)

	r1, ok := o.Get(k1)
	assert.True(t, ok)
	assert.Equal(t, v1, r1)

	o = o.Delete(k1)

	r2, ok := o.Get(k1)
	assert.False(t, ok)
	assert.Nil(t, r2)
}

func TestMap_Keys(t *testing.T) {
	k1 := NewString(faker.Word())
	v1 := NewString(faker.Word())

	o := NewMap(k1, v1)

	keys := o.Keys()
	assert.Len(t, keys, 1)
	assert.Contains(t, keys, k1)
}

func TestMap_Len(t *testing.T) {
	k1 := NewString(faker.Word())
	v1 := NewString(faker.Word())

	o1 := NewMap()
	o2 := NewMap(k1, v1)

	assert.Zero(t, o1.Len())
	assert.Equal(t, 1, o2.Len())
}

func TestMap_EncodeAndDecode(t *testing.T) {
	encoder := NewMapEncoder(NewStringEncoder())
	decoder := NewMapDecoder(NewStringDecoder())

	t.Run("Map", func(t *testing.T) {
		k1 := NewString(faker.Word())
		v1 := NewString(faker.Word())

		// Test Encode
		encoded, err := encoder.Encode(map[any]any{k1.Interface(): v1.Interface()})
		assert.NoError(t, err)
		assert.Equal(t, NewMap(k1, v1), encoded)

		// Test Decode
		var decoded map[any]any
		err = decoder.Decode(encoded, &decoded)
		assert.NoError(t, err)
		assert.Equal(t, map[any]any{k1.Interface(): v1.Interface()}, decoded)
	})

	t.Run("Struct", func(t *testing.T) {
		v1 := NewString(faker.Word())

		// Test Encode
		encoded, err := encoder.Encode(struct {
			K1 string
		}{
			K1: v1.String(),
		})
		assert.NoError(t, err)
		assert.True(t, NewMap(NewString("k_1"), v1).Compare(encoded) == 0)

		// Test Decode
		var decoded struct{ K1 string }
		err = decoder.Decode(encoded, &decoded)
		assert.NoError(t, err)
		assert.Equal(t, v1.String(), decoded.K1)
	})
}

func BenchmarkMap_Set(b *testing.B) {
	m := NewMap()

	for i := 0; i < b.N; i++ {
		m = m.Set(NewString(faker.Word()), NewString(faker.Word()))
	}
}

func BenchmarkMap_Get(b *testing.B) {
	m := NewMap()
	for i := 0; i < 1000; i++ {
		m = m.Set(NewString(faker.Word()), NewString(faker.Word()))
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = m.Get(NewString(faker.Word()))
	}
}

func BenchmarkMap_Interface(b *testing.B) {
	v := NewMap(NewString("key"), NewString("value"))

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = v.Interface()
	}
}
