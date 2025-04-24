package store

import (
	"testing"

	"github.com/go-faker/faker/v4"
	"github.com/stretchr/testify/require"
)

func TestSource_Open(t *testing.T) {
	src := NewSource()
	defer src.Close()

	name := faker.UUIDHyphenated()

	s1, err := src.Open(name)
	require.NoError(t, err)

	s2, err := src.Open(name)
	require.NoError(t, err)
	require.Equal(t, s1, s2)
}
