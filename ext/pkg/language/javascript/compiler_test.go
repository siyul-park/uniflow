package javascript

import (
	"context"
	"testing"
	"time"

	"github.com/go-faker/faker/v4"
	"github.com/stretchr/testify/assert"
)

func TestCompiler_Compile(t *testing.T) {
	c := NewCompiler()
	_, err := c.Compile(`export default function (args) {
		return args;
	}`)
	assert.NoError(t, err)
}

func TestProgram_Run(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.TODO(), time.Second)
	defer cancel()

	c := NewCompiler()
	p, _ := c.Compile(`export default function (args) {
		return args;
	}`)

	input := faker.UUIDHyphenated()

	output, err := p.Run(ctx, input)
	assert.NoError(t, err)
	assert.Equal(t, input, output)
}
