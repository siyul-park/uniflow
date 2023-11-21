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

func TestSlice_Hash(t *testing.T) {
	v1 := NewString(faker.Word())
	v2 := NewString(faker.Word())

	assert.NotEqual(t, NewSlice(v1, v2).Hash(), NewSlice(v2, v1).Hash())
	assert.Equal(t, NewSlice().Hash(), NewSlice().Hash())
	assert.Equal(t, NewSlice(v1, v2).Hash(), NewSlice(v1, v2).Hash())
}

func TestSlice_GetAndSet(t *testing.T) {
	v1 := NewString(faker.Word())
	v2 := NewString(faker.Word())

	o := NewSlice(v1)

	r1 := o.Get(0)
	assert.Equal(t, v1, r1)

	r2 := o.Get(1)
	assert.Nil(t, r2)

	o = o.Set(0, v2)

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

func TestSlice_Encode(t *testing.T) {
	e := NewSliceEncoder(NewStringEncoder())

	v1 := NewString(faker.Word())
	v2 := NewString(faker.Word())

	v, err := e.Encode([]string{v1.String(), v2.String()})
	assert.NoError(t, err)
	assert.Equal(t, NewSlice(v1, v2), v)
}

func TestSlice_Decode(t *testing.T) {
	d := NewSliceDecoder(NewStringDecoder())

	v1 := NewString(faker.Word())
	v2 := NewString(faker.Word())

	var v []string
	err := d.Decode(NewSlice(v1, v2), &v)
	assert.NoError(t, err)
	assert.Equal(t, []string{v1.String(), v2.String()}, v)
}