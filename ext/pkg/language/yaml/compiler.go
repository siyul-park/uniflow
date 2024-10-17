package yaml

import (
	"context"

	"gopkg.in/yaml.v3"

	"github.com/siyul-park/uniflow/ext/pkg/language"
)

const Language = "yaml"

func NewCompiler() language.Compiler {
	return language.CompileFunc(func(code string) (language.Program, error) {
		var data any
		if err := yaml.Unmarshal([]byte(code), &data); err != nil {
			return nil, err
		}
		return language.RunFunc(func(_ context.Context, _ []any) ([]any, error) {
			return []any{data}, nil
		}), nil
	})
}
