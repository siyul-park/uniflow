package node

import (
	"testing"

	"github.com/go-faker/faker/v4"
	"github.com/stretchr/testify/assert"
)

func TestPort_Format(t *testing.T) {
	name := faker.Word()
	index := 0

	port := PortWithIndex(name, index)

	n := NameOfPort(port)
	assert.Equal(t, name, n)

	i, ok := IndexOfPort(port)
	assert.True(t, ok)
	assert.Equal(t, index, i)
}
