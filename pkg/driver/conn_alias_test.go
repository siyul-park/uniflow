package driver

import (
	"testing"

	"github.com/go-faker/faker/v4"
	"github.com/stretchr/testify/require"
)

func TestConnAlias_Load(t *testing.T) {
	c := newConn()
	a := NewConnAlias(c)
	defer a.Close()

	name := faker.UUIDHyphenated()
	alias := faker.UUIDHyphenated()

	a.Alias(name, alias)

	s1, err := c.Load(name)
	require.NoError(t, err)
	require.NotNil(t, s1)

	s2, err := a.Load(alias)
	require.NoError(t, err)
	require.Equal(t, s1, s2)
}
