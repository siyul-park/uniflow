package text

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestCompiler_Compile(t *testing.T) {
	c := NewCompiler()
	_, err := c.Compile("foo")
	assert.NoError(t, err)
}

func TestProgram_Run(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.TODO(), time.Second)
	defer cancel()

	c := NewCompiler()
	p, _ := c.Compile("foo")

	output, err := p.Run(ctx)
	assert.NoError(t, err)
	assert.Equal(t, "foo", output)
}
