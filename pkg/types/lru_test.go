package types

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestLRU_Load(t *testing.T) {
	cache := NewLRU(2, 0)
	key1, value1 := NewBinary([]byte{1}), NewBinary([]byte{10})
	key2, value2 := NewBinary([]byte{2}), NewBinary([]byte{20})

	cache.Store(key1, value1)
	cache.Store(key2, value2)

	v, ok := cache.Load(key1)
	require.True(t, ok)
	require.Equal(t, value1, v)

	v, ok = cache.Load(key2)
	require.True(t, ok)
	require.Equal(t, value2, v)
}

func TestLRU_Store(t *testing.T) {
	cache := NewLRU(2, 0)
	key1, value1 := NewBinary([]byte{1}), NewBinary([]byte{10})
	key2, value2 := NewBinary([]byte{2}), NewBinary([]byte{20})

	cache.Store(key1, value1)
	cache.Store(key2, value2)

	v, ok := cache.Load(key1)
	require.True(t, ok)
	require.Equal(t, value1, v)

	v, ok = cache.Load(key2)
	require.True(t, ok)
	require.Equal(t, value2, v)

	cache.Store(key1, value2)

	v, ok = cache.Load(key1)
	require.True(t, ok)
	require.Equal(t, value2, v)
}

func TestLRU_Delete(t *testing.T) {
	cache := NewLRU(2, 0)
	key1, value1 := NewBinary([]byte{1}), NewBinary([]byte{10})
	key2, value2 := NewBinary([]byte{2}), NewBinary([]byte{20})

	cache.Store(key1, value1)
	cache.Store(key2, value2)

	cache.Delete(key1)
	v, ok := cache.Load(key1)
	require.False(t, ok)
	require.Nil(t, v)

	v, ok = cache.Load(key2)
	require.True(t, ok)
	require.Equal(t, value2, v)
}

func TestLRU_Evict(t *testing.T) {
	cache := NewLRU(2, 0)
	key1, value1 := NewBinary([]byte{1}), NewBinary([]byte{10})
	key2, value2 := NewBinary([]byte{2}), NewBinary([]byte{20})
	key3, value3 := NewBinary([]byte{3}), NewBinary([]byte{30})

	cache.Store(key1, value1)
	cache.Store(key2, value2)
	cache.Store(key3, value3)

	v, ok := cache.Load(key1)
	require.False(t, ok)
	require.Nil(t, v)

	v, ok = cache.Load(key2)
	require.True(t, ok)
	require.Equal(t, value2, v)

	v, ok = cache.Load(key3)
	require.True(t, ok)
	require.Equal(t, value3, v)
}

func TestLRU_Len(t *testing.T) {
	cache := NewLRU(2, 0)
	require.Equal(t, 0, cache.Len())

	key1, value1 := NewBinary([]byte{1}), NewBinary([]byte{10})
	cache.Store(key1, value1)
	require.Equal(t, 1, cache.Len())

	key2, value2 := NewBinary([]byte{2}), NewBinary([]byte{20})
	cache.Store(key2, value2)
	require.Equal(t, 2, cache.Len())

	key3, value3 := NewBinary([]byte{3}), NewBinary([]byte{30})
	cache.Store(key3, value3)
	require.Equal(t, 2, cache.Len())
}

func TestLRU_Clear(t *testing.T) {
	cache := NewLRU(2, 0)
	require.Equal(t, 0, cache.Len())

	key1, value1 := NewBinary([]byte{1}), NewBinary([]byte{10})
	cache.Store(key1, value1)
	require.Equal(t, 1, cache.Len())

	cache.Clear()
	require.Equal(t, 0, cache.Len())
}
