package js

import (
	"testing"

	"github.com/evanw/esbuild/pkg/api"
	"github.com/stretchr/testify/assert"
)

func TestTransform(t *testing.T) {
	ts := "let x: number = 1;"
	js, err := Transform(ts, api.TransformOptions{
		Loader: api.LoaderTS,
	})
	assert.NoError(t, err)
	assert.Equal(t, "let x = 1;\n", js)
}
