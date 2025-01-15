package text

import (
	"context"

	"github.com/siyul-park/uniflow/ext/pkg/language"
)

const Language = "text"

func NewCompiler() language.Compiler {
	return language.CompileFunc(func(code string) (language.Program, error) {
		return language.RunFunc(func(_ context.Context, _ ...any) (any, error) {
			return code, nil
		}), nil
	})
}
