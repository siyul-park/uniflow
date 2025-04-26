package scheme

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestSchemeBuilder_Register(t *testing.T) {
	b := NewBuilder()

	b.Register(RegisterFunc(func(_ *Scheme) error {
		return nil
	}))
	require.Equal(t, 1, b.Len())
}

func TestSchemeBuilder_AddToScheme(t *testing.T) {
	b := NewBuilder()

	b.Register(RegisterFunc(func(_ *Scheme) error {
		return nil
	}))

	err := b.AddToScheme(New())
	require.NoError(t, err)
}

func TestSchemeBuilder_Build(t *testing.T) {
	b := NewBuilder()

	b.Register(RegisterFunc(func(_ *Scheme) error {
		return nil
	}))

	s, err := b.Build()
	require.NoError(t, err)
	require.NotNil(t, s)
}
