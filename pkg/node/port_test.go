package node

import (
	"testing"

	"github.com/go-faker/faker/v4"
	"github.com/stretchr/testify/require"
)

func TestPort_Format(t *testing.T) {
	name := faker.Word()
	index := 0

	port := PortWithIndex(name, index)

	n := NameOfPort(port)
	require.Equal(t, name, n)

	i, ok := IndexOfPort(port)
	require.True(t, ok)
	require.Equal(t, index, i)
}
