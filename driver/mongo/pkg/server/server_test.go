package server

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewAndRelease(t *testing.T) {
	s := New()
	assert.NotNil(t, s)

	Release(s)
}
