package pool

import (
	"github.com/go-faker/faker/v4"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestGetMap(t *testing.T) {
	m := GetMap()
	assert.NotNil(t, m)
}

func TestPutMap(t *testing.T) {
	m := GetMap()

	m.Store(faker.UUIDHyphenated(), faker.UUIDHyphenated())

	PutMap(m)

	count := 0
	m.Range(func(_, _ any) bool {
		count += 1
		return true
	})

	assert.Equal(t, 0, count)
}
