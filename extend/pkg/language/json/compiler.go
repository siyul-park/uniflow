package json

import (
	"encoding/json"

	"github.com/siyul-park/uniflow/extend/pkg/language"
)

const Language = "json"

func NewCompiler() language.Compiler {
	return language.CompileFunc(func(code string) (language.Program, error) {
		var data any
		if err := json.Unmarshal([]byte(code), &data); err != nil {
			return nil, err
		}
		return language.RunFunc(func(_ any) (any, error) {
			return data, nil
		}), nil
	})
}
