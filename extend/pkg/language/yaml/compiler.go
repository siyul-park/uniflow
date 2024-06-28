package yaml

import (
	"gopkg.in/yaml.v3"

	"github.com/siyul-park/uniflow/extend/pkg/language"
)

func NewCompiler() language.Compiler {
	return language.CompileFunc(func(code string) (language.Program, error) {
		var data any
		if err := yaml.Unmarshal([]byte(code), &data); err != nil {
			return nil, err
		}
		return language.RunFunc(func(_ any) (any, error) {
			return data, nil
		}), nil
	})
}
