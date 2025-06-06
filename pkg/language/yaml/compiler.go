package yaml

import (
	"context"

	"gopkg.in/yaml.v3"

	"github.com/siyul-park/uniflow/pkg/language"
)

const Language = "yaml"

// NewCompiler returns a compiler that processes YAML code.
func NewCompiler() language.Compiler {
	return language.CompileFunc(func(code string) (language.Program, error) {
		var data any
		if err := yaml.Unmarshal([]byte(code), &data); err != nil {
			return nil, err
		}
		return language.RunFunc(func(_ context.Context, _ ...any) (any, error) {
			return data, nil
		}), nil
	})
}
