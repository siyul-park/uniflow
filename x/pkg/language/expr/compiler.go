package expr

import (
	"github.com/expr-lang/expr"
	"github.com/siyul-park/uniflow/x/pkg/language"
)

func NewCompiler() language.Compiler {
	return language.CompileFunc(func(code string) (language.Program, error) {
		program, err := expr.Compile(code)
		if err != nil {
			return nil, err
		}
		return language.RunFunc(func(env any) (any, error) {
			return expr.Run(program, env)
		}), nil
	})
}
