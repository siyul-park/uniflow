package node

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestUnwrap(t *testing.T) {
	n := NewOneToOneNode(nil)
	defer n.Close()

	p := NoCloser(n)
	assert.Equal(t, n, Unwrap(p))
}
