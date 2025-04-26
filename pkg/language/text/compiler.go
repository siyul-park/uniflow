package text

import (
	"context"

	"github.com/siyul-park/uniflow/pkg/language"
)

const Language = "text"

// NewCompiler returns a compiler that handles plain text code.
func NewCompiler() language.Compiler {
	return language.CompileFunc(func(code string) (language.Program, error) {
		return language.RunFunc(func(_ context.Context, _ ...any) (any, error) {
			return code, nil
		}), nil
	})
}
