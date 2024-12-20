package node

import (
	"github.com/stretchr/testify/assert"
	"testing"
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

func TestNoCloser_Close(t *testing.T) {
	n := NewOneToOneNode(nil)
	defer n.Close()

	p := NoCloser(n)
	assert.NoError(t, p.Close())
}
