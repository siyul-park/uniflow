package server

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestNewAndRelease(t *testing.T) {
	s := New()
	require.NotNil(t, s)

	Release(s)
}
