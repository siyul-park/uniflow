package cel

import (
	"context"

	"github.com/google/cel-go/cel"
	"github.com/google/cel-go/ext"
	"github.com/siyul-park/uniflow/ext/pkg/language"
)

const Language = "cel"

func NewCompiler(opts ...cel.EnvOption) language.Compiler {
	opts = append(opts, cel.CustomTypeAdapter(&adapter{}), ext.Encoders(), ext.Math(), ext.Lists(), ext.Sets(), ext.Strings(), cel.Variable("self", cel.AnyType))
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

		return language.RunFunc(func(ctx context.Context, args []any) ([]any, error) {
			var env any
			if len(args) == 0 {
				env = nil
			} else if len(args) == 1 {
				env = args[0]
			} else {
				env = args
			}

			val, _, err := prg.ContextEval(ctx, map[string]any{
				"self": env,
			})
			if err != nil {
				return nil, err
			}
			return []any{val.Value()}, nil
		}), nil
	})
}
