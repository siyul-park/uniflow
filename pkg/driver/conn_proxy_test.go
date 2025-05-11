package driver

import (
	"testing"

	"github.com/go-faker/faker/v4"
	"github.com/stretchr/testify/require"
)

func TestConnProxy_Load(t *testing.T) {
	c := newConn()
	p := NewConnProxy(c)
	defer p.Close()

	name := faker.UUIDHyphenated()

	s1, err := p.Load(name)
	require.NoError(t, err)
	require.NotNil(t, s1)

	s2, err := p.Load(name)
	require.NoError(t, err)
	require.Equal(t, s1, s2)
}

func TestConnProxy_Wrap(t *testing.T) {
	c := newConn()
	p := NewConnProxy(nil)
	defer p.Close()

	p.Wrap(c)

	r := p.Unwrap()
	require.Equal(t, c, r)
}

func TestConnProxy_Unwrap(t *testing.T) {
	c := newConn()
	p := NewConnProxy(c)
	defer p.Close()

	r := p.Unwrap()
	require.Equal(t, c, r)
}

func TestConnProxy_Close(t *testing.T) {
	c := newConn()
	p := NewConnProxy(c)

	err := p.Close()
	require.NoError(t, err)
}
