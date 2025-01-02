package node

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNoCloser(t *testing.T) {
	n := NewOneToOneNode(nil)
	defer n.Close()

	p := NoCloser(n)
	assert.Equal(t, n, Unwrap(p))
	assert.NoError(t, p.Close())
}

func TestUnwrap(t *testing.T) {
	t.Run("NoProxy", func(t *testing.T) {
		n := NewOneToOneNode(nil)
		defer n.Close()

		assert.Equal(t, n, Unwrap(n))
	})

	t.Run("ProxyNode", func(t *testing.T) {
		n := NewOneToOneNode(nil)
		defer n.Close()

		p := NoCloser(n)
		assert.Equal(t, n, Unwrap(p))
	})
}

func TestAs(t *testing.T) {
	t.Run("NoProxy", func(t *testing.T) {
		n := NewOneToOneNode(nil)
		defer n.Close()

		var target *OneToOneNode
		assert.True(t, As(n, &target))
		assert.Equal(t, n, target)
	})

	t.Run("ProxyNode", func(t *testing.T) {
		n := NewOneToOneNode(nil)
		defer n.Close()

		p := NoCloser(n)
		var target *OneToOneNode
		assert.True(t, As(p, &target))
		assert.Equal(t, n, target)
	})
}

func TestNoCloser_Close(t *testing.T) {
	n := NewOneToOneNode(nil)
	defer n.Close()

	p := NoCloser(n)
	assert.NoError(t, p.Close())
}
