package process

import (
	"testing"

	"github.com/go-faker/faker/v4"
	"github.com/stretchr/testify/assert"
)

func TestData_LoadAndDelete(t *testing.T) {
	d := newData()
	defer d.Close()

	k := faker.UUIDHyphenated()
	v := faker.UUIDHyphenated()

	d.Store(k, v)

	r := d.LoadAndDelete(k)
	assert.Equal(t, v, r)

	r = d.Load(k)
	assert.Equal(t, nil, r)
}

func TestData_Load(t *testing.T) {
	d := newData()
	defer d.Close()

	k := faker.UUIDHyphenated()
	v := faker.UUIDHyphenated()

	r := d.Load(k)
	assert.Equal(t, nil, r)

	d.Store(k, v)

	r = d.Load(k)
	assert.Equal(t, v, r)
}

func TestData_Store(t *testing.T) {
	d := newData()
	defer d.Close()

	k := faker.UUIDHyphenated()
	v1 := faker.UUIDHyphenated()
	v2 := faker.UUIDHyphenated()

	d.Store(k, v1)
	d.Store(k, v2)

	r := d.Load(k)
	assert.Equal(t, v2, r)
}

func TestData_Delete(t *testing.T) {
	d := newData()
	defer d.Close()

	k := faker.UUIDHyphenated()
	v := faker.UUIDHyphenated()

	ok := d.Delete(k)
	assert.False(t, ok)

	d.Store(k, v)

	ok = d.Delete(k)
	assert.True(t, ok)
}

func TestData_Fork(t *testing.T) {
	d := newData()
	defer d.Close()

	c := d.Fork()

	k := faker.UUIDHyphenated()
	v1 := faker.UUIDHyphenated()
	v2 := faker.UUIDHyphenated()

	d.Store(k, v1)

	r := c.Load(k)
	assert.Equal(t, v1, r)

	c.Store(k, v2)

	r = c.Load(k)
	assert.Equal(t, v2, r)

	ok := c.Delete(k)
	assert.True(t, ok)

	ok = d.Delete(k)
	assert.True(t, ok)
}
