package hook

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestHooksBuilder_Register(t *testing.T) {
	b := NewBuilder()

	b.Register(RegisterFunc(func(_ *Hook) error {
		return nil
	}))
	assert.Len(t, b, 1)
}

func TestHooksBuilder_AddToScheme(t *testing.T) {
	b := NewBuilder()

	b.Register(RegisterFunc(func(_ *Hook) error {
		return nil
	}))

	err := b.AddToHooks(New())
	assert.NoError(t, err)
}
