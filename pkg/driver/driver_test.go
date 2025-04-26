package driver

import (
	"testing"

	"github.com/go-faker/faker/v4"
	"github.com/stretchr/testify/require"
)

func TestDriver_Open(t *testing.T) {
	d := New()
	defer d.Close()

	name := faker.UUIDHyphenated()

	c1, err := d.Open(name)
	require.NoError(t, err)
	require.NotNil(t, c1)

	c2, err := d.Open(name)
	require.NoError(t, err)
	require.Equal(t, c1, c2)
}
