package cel

import (
	"context"
	"reflect"

	"github.com/google/cel-go/cel"
	"github.com/siyul-park/uniflow/pkg/language"
)

const Language = "ecmascript"

// NewCompiler creates and returns a new language.Compiler for CEL with the specified options.
func NewCompiler(opts ...cel.EnvOption) language.Compiler {
	opts = append(opts, cel.CustomTypeAdapter(TypeAdapter), cel.StdLib())
	return language.CompileFunc(func(code string) (language.Program, error) {
		env, err := cel.NewEnv(opts...)
		if err != nil {
			return nil, err
		}
		ast, issues := env.Parse(code)
		if issues != nil && issues.Err() != nil {
			return nil, issues.Err()
		}
		prg, err := env.Program(ast)
		if err != nil {
			return nil, err
		}

		return language.RunFunc(func(ctx context.Context, args ...any) (any, error) {
			env := map[string]any{}
			if len(args) == 1 {
				self := reflect.ValueOf(args[0])
				if self.Kind() == reflect.Map {
					env = map[string]any{}
					for _, key := range self.MapKeys() {
						env[key.String()] = self.MapIndex(key).Interface()
					}
				}
				env["self"] = args[0]
			} else {
				env = map[string]any{"self": args}
			}

			val, _, err := prg.ContextEval(ctx, env)
			if err != nil {
				return nil, err
			}
			return val.Value(), nil
		}), nil
	})
}
