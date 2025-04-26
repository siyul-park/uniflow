package language

import (
	"context"
	"testing"

	"github.com/go-faker/faker/v4"
	"github.com/stretchr/testify/require"
)

func TestCompileFunc(t *testing.T) {
	c := CompileFunc(func(_ string) (Program, error) {
		return RunFunc(func(_ context.Context, _ ...any) (any, error) {
			return nil, nil
		}), nil
	})

	code := faker.Paragraph()

	p, err := c.Compile(code)
	require.NoError(t, err)
	require.NotNil(t, p)
}
