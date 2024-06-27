package jsonata

import (
	"testing"

	"github.com/go-faker/faker/v4"
	"github.com/stretchr/testify/assert"
)

func TestCompiler_Compile(t *testing.T) {
	c := NewCompiler()
	_, err := c.Compile("$")
	assert.NoError(t, err)
}

func TestProgram_Run(t *testing.T) {
	c := NewCompiler()
	p, _ := c.Compile("$")

	env := faker.Word()

	res, err := p.Run(env)
	assert.NoError(t, err)
	assert.Equal(t, env, res)
}
