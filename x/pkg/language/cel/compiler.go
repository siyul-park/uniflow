package cel

import (
	"github.com/google/cel-go/cel"
	"github.com/siyul-park/uniflow/x/pkg/language"
)

func NewCompiler() language.Compiler {
	return language.CompileFunc(func(code string) (language.Program, error) {
		env, err := cel.NewEnv(
			cel.Variable("self", cel.AnyType),
		)
		if err != nil {
			return nil, err
		}
		ast, issues := env.Compile(code)
		if issues != nil && issues.Err() != nil {
			return nil, issues.Err()
		}
		prg, err := env.Program(ast)
		if err != nil {
			return nil, err
		}
		return language.RunFunc(func(env any) (any, error) {
			val, _, err := prg.Eval(map[string]any{
				"self": env,
			})
			if err != nil {
				return nil, err
			}
			return val.Value(), nil
		}), nil
	})
}
