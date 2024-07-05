package process

import (
	"testing"

	"github.com/go-faker/faker/v4"
	"github.com/stretchr/testify/assert"
)

func TestLocal_Watch(t *testing.T) {
	l := NewLocal[string]()
	defer l.Close()

	proc := New()
	defer proc.Exit(nil)

	count := 0
	ok := l.Watch(proc, func(_ string) {
		count++
	})
	assert.False(t, ok)

	v := faker.UUIDHyphenated()

	l.Store(proc, v)
	assert.Equal(t, 1, count)

	ok = l.Watch(proc, func(_ string) {
		count++
	})
	assert.True(t, ok)
	assert.Equal(t, 2, count)
}

func TestLocal_Keys(t *testing.T) {
	l := NewLocal[string]()
	defer l.Close()

	proc := New()
	defer proc.Exit(nil)

	v := faker.UUIDHyphenated()

	l.Store(proc, v)

	keys := l.Keys()
	assert.Contains(t, keys, proc)
}

func TestLocal_Load(t *testing.T) {
	l := NewLocal[string]()
	defer l.Close()

	proc := New()
	defer proc.Exit(nil)

	v := faker.UUIDHyphenated()

	_, ok := l.Load(proc)
	assert.False(t, ok)

	l.Store(proc, v)

	r, ok := l.Load(proc)
	assert.True(t, ok)
	assert.Equal(t, v, r)
}

func TestLocal_Store(t *testing.T) {
	l := NewLocal[string]()
	defer l.Close()

	proc := New()
	defer proc.Exit(nil)

	v1 := faker.UUIDHyphenated()
	v2 := faker.UUIDHyphenated()

	l.Store(proc, v1)
	l.Store(proc, v2)

	r, ok := l.Load(proc)
	assert.True(t, ok)
	assert.Equal(t, v2, r)
}

func TestLocal_Delete(t *testing.T) {
	l := NewLocal[string]()
	defer l.Close()

	proc := New()
	defer proc.Exit(nil)

	v := faker.UUIDHyphenated()

	ok := l.Delete(proc)
	assert.False(t, ok)

	l.Store(proc, v)

	ok = l.Delete(proc)
	assert.True(t, ok)
}

func TestLocal_LoadOrStore(t *testing.T) {
	l := NewLocal[string]()
	defer l.Close()

	proc := New()
	defer proc.Exit(nil)

	v := faker.UUIDHyphenated()

	r, err := l.LoadOrStore(proc, func() (string, error) {
		return v, nil
	})
	assert.NoError(t, err)
	assert.Equal(t, v, r)
}
