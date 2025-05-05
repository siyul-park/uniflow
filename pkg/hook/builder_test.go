package hook

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestHooksBuilder_Register(t *testing.T) {
	b := NewBuilder()

	r := RegisterFunc(func(_ *Hook) error {
		return nil
	})

	ok := b.Register(r)
	require.True(t, ok)
	require.Equal(t, 1, b.Len())

	ok = b.Register(r)
	require.False(t, ok)
	require.Equal(t, 1, b.Len())
}

func TestHooksBuilder_Unregister(t *testing.T) {
	b := NewBuilder()

	r := RegisterFunc(func(_ *Hook) error {
		return nil
	})

	ok := b.Register(r)
	require.True(t, ok)
	require.Equal(t, 1, b.Len())

	ok = b.Unregister(r)
	require.True(t, ok)
	require.Equal(t, 0, b.Len())
}

func TestHooksBuilder_AddToScheme(t *testing.T) {
	b := NewBuilder()

	b.Register(RegisterFunc(func(_ *Hook) error {
		return nil
	}))

	err := b.AddToHook(New())
	require.NoError(t, err)
}
