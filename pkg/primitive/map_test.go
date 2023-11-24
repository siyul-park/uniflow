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

	assert.Len(t, o.Keys(), 1)
	assert.Contains(t, o.Keys(), k1)
}

func TestMap_Len(t *testing.T) {
	k1 := NewString(faker.Word())
	v1 := NewString(faker.Word())

	o1 := NewMap()
	o2 := NewMap(k1, v1)

	assert.Equal(t, 0, o1.Len())
	assert.Equal(t, 1, o2.Len())
}

func TestMap_Encode(t *testing.T) {
	e := NewMapEncoder(NewStringEncoder())

	t.Run("map", func(t *testing.T) {
		k1 := NewString(faker.Word())
		v1 := NewString(faker.Word())

		v, err := e.Encode(map[string]string{k1.String(): v1.String()})
		assert.NoError(t, err)
		assert.Equal(t, NewMap(k1, v1), v)
	})

	t.Run("struct", func(t *testing.T) {
		v1 := NewString(faker.Word())

		v, err := e.Encode(struct {
			K1 string
		}{
			K1: v1.String(),
		})
		assert.NoError(t, err)
		assert.True(t, NewMap(NewString("k_1"), v1).Compare(v) == 0)
	})
}

func TestMap_Decode(t *testing.T) {
	d := NewMapDecoder(NewStringDecoder())

	t.Run("map", func(t *testing.T) {
		k1 := NewString(faker.Word())
		v1 := NewString(faker.Word())

		var v map[string]string
		err := d.Decode(NewMap(k1, v1), &v)
		assert.NoError(t, err)
		assert.Equal(t, map[string]string{k1.String(): v1.String()}, v)
	})
	t.Run("struct", func(t *testing.T) {
		v1 := NewString(faker.Word())

		var v struct{ K1 string }
		err := d.Decode(NewMap(NewString("k_1"), v1), &v)
		assert.NoError(t, err)
		assert.Equal(t, v1.String(), v.K1)
	})
}
