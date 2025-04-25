package process

import (
	"testing"

	"github.com/go-faker/faker/v4"
	"github.com/stretchr/testify/require"
)

func TestLocal_AddStoreHook(t *testing.T) {
	l := NewLocal[string]()
	defer l.Close()

	proc := New()
	defer proc.Exit(nil)

	count := 0
	h := StoreFunc(func(_ string) {
		count++
	})

	ok := l.AddStoreHook(proc, h)
	require.True(t, ok)

	v := faker.UUIDHyphenated()

	l.Store(proc, v)
	require.Equal(t, 1, count)

	ok = l.RemoveStoreHook(proc, h)
	require.False(t, ok)

	ok = l.AddStoreHook(proc, h)
	require.True(t, ok)
	require.Equal(t, 2, count)

	ok = l.RemoveStoreHook(proc, h)
	require.False(t, ok)
}

func TestLocal_Keys(t *testing.T) {
	l := NewLocal[string]()
	defer l.Close()

	proc := New()
	defer proc.Exit(nil)

	v := faker.UUIDHyphenated()

	l.Store(proc, v)

	keys := l.Keys()
	require.Contains(t, keys, proc)
}

func TestLocal_Load(t *testing.T) {
	l := NewLocal[string]()
	defer l.Close()

	proc := New()
	defer proc.Exit(nil)

	v := faker.UUIDHyphenated()

	_, ok := l.Load(proc)
	require.False(t, ok)

	l.Store(proc, v)

	r, ok := l.Load(proc)
	require.True(t, ok)
	require.Equal(t, v, r)
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
	require.True(t, ok)
	require.Equal(t, v2, r)
}

func TestLocal_Delete(t *testing.T) {
	l := NewLocal[string]()
	defer l.Close()

	proc := New()
	defer proc.Exit(nil)

	v := faker.UUIDHyphenated()

	ok := l.Delete(proc)
	require.False(t, ok)

	l.Store(proc, v)

	ok = l.Delete(proc)
	require.True(t, ok)
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
	require.NoError(t, err)
	require.Equal(t, v, r)
}
