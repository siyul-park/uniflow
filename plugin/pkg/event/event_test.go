package event

import (
	"testing"

	"github.com/go-faker/faker/v4"
	"github.com/stretchr/testify/assert"
)

func TestNew(t *testing.T) {
	d := faker.UUIDHyphenated()
	e := New(d)
	assert.Equal(t, d, e.Data())
}
