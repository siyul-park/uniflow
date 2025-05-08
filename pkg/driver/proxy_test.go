package driver

import (
	"testing"

	"github.com/go-faker/faker/v4"
	"github.com/stretchr/testify/require"
)

func TestProxy_Open(t *testing.T) {
	d := New()
	p := NewProxy(d)
	defer p.Close()

	name := faker.UUIDHyphenated()

	c1, err := p.Open(name)
	require.NoError(t, err)
	require.NotNil(t, c1)

	c2, err := p.Open(name)
	require.NoError(t, err)
	require.Equal(t, c1, c2)
}

func TestProxy_Wrap(t *testing.T) {
	d := New()
	p := NewProxy(nil)
	defer p.Close()

	p.Wrap(d)

	r := p.Unwrap()
	require.Equal(t, d, r)
}

func TestProxy_Unwrap(t *testing.T) {
	d := New()
	p := NewProxy(d)
	defer p.Close()

	r := p.Unwrap()
	require.Equal(t, d, r)
}

func TestProxy_Close(t *testing.T) {
	d := New()
	p := NewProxy(d)
	defer p.Close()

	err := p.Close()
	require.NoError(t, err)
}
