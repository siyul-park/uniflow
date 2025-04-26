package driver

import (
	"testing"

	"github.com/go-faker/faker/v4"
	"github.com/stretchr/testify/require"
)

func TestRegistry_Lookup(t *testing.T) {
	reg := NewRegistry()
	defer reg.Close()

	name := faker.UUIDHyphenated()

	drv := New()

	err := reg.Register(name, drv)
	require.NoError(t, err)

	d, err := reg.Lookup(name)
	require.NoError(t, err)
	require.Equal(t, d, drv)
}
