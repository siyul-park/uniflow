package pipe

import (
	"testing"

	"github.com/go-faker/faker/v4"
	"github.com/stretchr/testify/assert"
)

func TestWritePipe_Write(t *testing.T) {
	p0 := newWrite[string]()

	p1 := newRead[string](0)
	defer p1.Close()

	p0.Link(p1)

	data1 := faker.UUIDHyphenated()
	data2 := faker.UUIDHyphenated()

	p0.Write(data1)
	p0.Write(data2)

	assert.Equal(t, data1, <-p1.Read())
	assert.Equal(t, data2, <-p1.Read())
}
