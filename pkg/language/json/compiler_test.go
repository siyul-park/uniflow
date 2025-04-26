package json

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestCompiler_Compile(t *testing.T) {
	c := NewCompiler()
	_, err := c.Compile("\"foo\"")
	require.NoError(t, err)
}

func TestProgram_Run(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.TODO(), time.Second)
	defer cancel()

	c := NewCompiler()
	p, _ := c.Compile("\"foo\"")

	output, err := p.Run(ctx)
	require.NoError(t, err)
	require.Equal(t, "foo", output)
}
