package cel

import (
	"github.com/google/cel-go/cel"
	"github.com/google/cel-go/ext"
	"github.com/siyul-park/uniflow/extend/pkg/language"
)

const Language = "cel"

func NewCompiler(opts ...cel.EnvOption) language.Compiler {
	opts = append(opts, ext.Encoders(), ext.Math(), ext.Lists(), ext.Sets(), ext.Strings(), cel.Variable("self", cel.AnyType))
	return language.CompileFunc(func(code string) (language.Program, error) {
		env, err := cel.NewEnv(opts...)
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
