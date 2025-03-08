package javascript

import (
	"context"
	"testing"
	"time"

	"github.com/go-faker/faker/v4"
	"github.com/stretchr/testify/require"
)

func TestCompiler_Compile(t *testing.T) {
	c := NewCompiler()
	_, err := c.Compile(`export default function (args) {
		return args;
	}`)
	require.NoError(t, err)
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
	require.NoError(t, err)
	require.Equal(t, input, output)
}
