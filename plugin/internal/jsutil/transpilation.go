package jsutil

import (
	"errors"
	"strings"

	"github.com/evanw/esbuild/pkg/api"
)

func Transform(code string, options api.TransformOptions) (string, error) {
	if result := api.Transform(code, options); len(result.Errors) > 0 {
		var msgs []string
		for _, err := range result.Errors {
			msgs = append(msgs, err.Text)
		}
		return "", errors.New(strings.Join(msgs, ", "))
	} else {
		return string(result.Code), nil
	}
}
