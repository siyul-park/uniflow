package driver

import (
	"testing"

	"github.com/go-faker/faker/v4"
	"github.com/stretchr/testify/require"
)

func TestConn_Load(t *testing.T) {
	c := newConn()
	defer c.Close()

	name := faker.UUIDHyphenated()

	s1, err := c.Load(name)
	require.NoError(t, err)
	require.NotNil(t, s1)

	s2, err := c.Load(name)
	require.NoError(t, err)
	require.Equal(t, s1, s2)
}
