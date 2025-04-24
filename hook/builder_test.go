package hook

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestHooksBuilder_Register(t *testing.T) {
	b := NewBuilder()

	b.Register(RegisterFunc(func(_ *Hook) error {
		return nil
	}))
	require.Len(t, b, 1)
}

func TestHooksBuilder_AddToScheme(t *testing.T) {
	b := NewBuilder()

	b.Register(RegisterFunc(func(_ *Hook) error {
		return nil
	}))

	err := b.AddToHook(New())
	require.NoError(t, err)
}
