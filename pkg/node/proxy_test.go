package node

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestNoCloser(t *testing.T) {
	n := NewOneToOneNode(nil)
	defer n.Close()

	p := NoCloser(n)
	require.Equal(t, n, Unwrap(p))
	require.NoError(t, p.Close())
}

func TestUnwrap(t *testing.T) {
	t.Run("NoProxy", func(t *testing.T) {
		n := NewOneToOneNode(nil)
		defer n.Close()

		require.Nil(t, Unwrap(n))
	})

	t.Run("ProxyNode", func(t *testing.T) {
		n := NewOneToOneNode(nil)
		defer n.Close()

		p := NoCloser(n)
		require.Equal(t, n, Unwrap(p))
	})
}

func TestAs(t *testing.T) {
	t.Run("NoProxy", func(t *testing.T) {
		n := NewOneToOneNode(nil)
		defer n.Close()

		var target *OneToOneNode
		require.True(t, As(n, &target))
		require.Equal(t, n, target)
	})

	t.Run("ProxyNode", func(t *testing.T) {
		n := NewOneToOneNode(nil)
		defer n.Close()

		p := NoCloser(n)
		var target *OneToOneNode
		require.True(t, As(p, &target))
		require.Equal(t, n, target)
	})
}

func TestNoCloser_Close(t *testing.T) {
	n := NewOneToOneNode(nil)
	defer n.Close()

	p := NoCloser(n)
	require.NoError(t, p.Close())
}
