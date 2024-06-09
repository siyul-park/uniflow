package spec

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSchemeBuilder_Register(t *testing.T) {
	b := NewBuilder()

	b.Register(func(_ *Scheme) error { return nil })
	assert.Len(t, b, 1)
}

func TestSchemeBuilder_AddToScheme(t *testing.T) {
	b := NewBuilder()

	b.Register(func(_ *Scheme) error { return nil })

	err := b.AddToScheme(NewScheme())
	assert.NoError(t, err)
}

func TestSchemeBuilder_Build(t *testing.T) {
	b := NewBuilder()

	b.Register(func(_ *Scheme) error { return nil })

	s, err := b.Build()
	assert.NoError(t, err)
	assert.NotNil(t, s)
}
