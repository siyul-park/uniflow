package port

import (
	"testing"

	"github.com/go-faker/faker/v4"
	"github.com/stretchr/testify/assert"
)

func TestGetAndSetIndexInPort(t *testing.T) {
	port := faker.UUIDHyphenated()
	index := 0

	i, ok := GetIndex(port, SetIndex(port, index))
	assert.True(t, ok)
	assert.Equal(t, index, i)
}
