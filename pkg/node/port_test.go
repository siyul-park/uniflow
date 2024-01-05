package node

import (
	"testing"

	"github.com/go-faker/faker/v4"
	"github.com/stretchr/testify/assert"
)

func TestNewMultiPortAndGetIndex(t *testing.T) {
	port := faker.UUIDHyphenated()
	index := 0

	i, ok := IndexOfMultiPort(port, MultiPort(port, index))
	assert.True(t, ok)
	assert.Equal(t, index, i)
}
