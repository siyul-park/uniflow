package network

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewRouteNode(t *testing.T) {
	n := NewRouteNode()
	assert.NotNil(t, n)
}
