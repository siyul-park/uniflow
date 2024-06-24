package javascript

import (
	"testing"

	"github.com/evanw/esbuild/pkg/api"
	"github.com/go-faker/faker/v4"
	"github.com/stretchr/testify/assert"
)

func TestCompiler_Compile(t *testing.T) {
	c := NewCompiler(api.TransformOptions{})
	_, err := c.Compile(`export default function (msg) {
		return msg;
	}`)
	assert.NoError(t, err)
}

func TestProgram_Run(t *testing.T) {
	c := NewCompiler(api.TransformOptions{})
	p, _ := c.Compile(`export default function (msg) {
		return msg;
	}`)

	env := faker.Word()

	res, err := p.Run(env)
	assert.NoError(t, err)
	assert.Equal(t, env, res)
}
