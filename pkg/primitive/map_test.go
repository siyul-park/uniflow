package primitive

import (
	"testing"

	"github.com/go-faker/faker/v4"
	"github.com/stretchr/testify/assert"
)

func TestNewMap(t *testing.T) {
	k1 := NewString(faker.UUIDHyphenated())
	v1 := NewString(faker.UUIDHyphenated())

	o := NewMap(k1, v1)

	assert.Equal(t, KindMap, o.Kind())
	assert.Equal(t, map[string]string{k1.String(): v1.String()}, o.Interface())
	assert.Equal(t, map[any]any{k1.Interface(): v1.Interface()}, o.Map())
}

func TestMap_GetAndSetAndDelete(t *testing.T) {
	k1 := NewString(faker.UUIDHyphenated())
	v1 := NewString(faker.UUIDHyphenated())

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
	k1 := NewString(faker.UUIDHyphenated())
	v1 := NewString(faker.UUIDHyphenated())

	o := NewMap(k1, v1)

	keys := o.Keys()
	assert.Len(t, keys, 1)
	assert.Contains(t, keys, k1)
}

func TestMap_Values(t *testing.T) {
	k1 := NewString(faker.UUIDHyphenated())
	v1 := NewString(faker.UUIDHyphenated())

	o := NewMap(k1, v1)

	values := o.Values()
	assert.Len(t, values, 1)
	assert.Contains(t, values, v1)
}

func TestMap_Pairs(t *testing.T) {
	k1 := NewString(faker.UUIDHyphenated())
	v1 := NewString(faker.UUIDHyphenated())

	o := NewMap(k1, v1)

	pairs := o.Pairs()
	assert.Len(t, pairs, 2)
	assert.Contains(t, pairs, k1)
	assert.Contains(t, pairs, v1)
}

func TestMap_Len(t *testing.T) {
	k1 := NewString(faker.UUIDHyphenated())
	v1 := NewString(faker.UUIDHyphenated())

	o1 := NewMap()
	o2 := NewMap(k1, v1)

	assert.Zero(t, o1.Len())
	assert.Equal(t, 1, o2.Len())
}

func TestMap_EncodeAndDecode(t *testing.T) {
	encoder := newMapEncoder(newStringEncoder())
	decoder := newMapDecoder(newStringDecoder())

	t.Run("Map", func(t *testing.T) {
		k1 := NewString(faker.UUIDHyphenated())
		v1 := NewString(faker.UUIDHyphenated())

		encoded, err := encoder.Encode(map[any]any{k1.Interface(): v1.Interface()})
		assert.NoError(t, err)
		assert.Equal(t, NewMap(k1, v1), encoded)

		var decoded map[any]any
		err = decoder.Decode(encoded, &decoded)
		assert.NoError(t, err)
		assert.Equal(t, map[any]any{k1.Interface(): v1.Interface()}, decoded)
	})

	t.Run("Struct", func(t *testing.T) {
		v1 := NewString(faker.UUIDHyphenated())

		encoded, err := encoder.Encode(struct {
			K1 string
		}{
			K1: v1.String(),
		})
		assert.NoError(t, err)
		assert.True(t, NewMap(NewString("k_1"), v1).Compare(encoded) == 0)

		var decoded struct{ K1 string }
		err = decoder.Decode(encoded, &decoded)
		assert.NoError(t, err)
		assert.Equal(t, v1.String(), decoded.K1)
	})
}

func BenchmarkMap_Set(b *testing.B) {
	m := NewMap()

	for i := 0; i < b.N; i++ {
		m = m.Set(NewString(faker.UUIDHyphenated()), NewString(faker.UUIDHyphenated()))
	}
}

func BenchmarkMap_Get(b *testing.B) {
	m := NewMap()
	for i := 0; i < 1000; i++ {
		m = m.Set(NewString(faker.UUIDHyphenated()), NewString(faker.UUIDHyphenated()))
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = m.Get(NewString(faker.UUIDHyphenated()))
	}
}

func BenchmarkMap_Interface(b *testing.B) {
	v := NewMap(NewString("key"), NewString("value"))

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = v.Interface()
	}
}

func BenchmarkMap_EncodeAndDecode(b *testing.B) {
	encoder := newMapEncoder(newStringEncoder())
	decoder := newMapDecoder(newStringDecoder())

	b.Run("Map", func(b *testing.B) {
		k1 := NewString(faker.UUIDHyphenated())
		v1 := NewString(faker.UUIDHyphenated())

		for i := 0; i < b.N; i++ {
			encoded, _ := encoder.Encode(map[any]any{k1.Interface(): v1.Interface()})

			var decoded map[any]any
			_ = decoder.Decode(encoded, &decoded)
		}
	})

	b.Run("Struct", func(b *testing.B) {
		v1 := NewString(faker.UUIDHyphenated())

		for i := 0; i < b.N; i++ {
			encoded, _ := encoder.Encode(struct {
				K1 string
			}{
				K1: v1.String(),
			})

			var decoded struct{ K1 string }
			_ = decoder.Decode(encoded, &decoded)
		}
	})
}
