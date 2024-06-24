package jsonata

import (
	"github.com/siyul-park/uniflow/plugin/pkg/language"
	"github.com/xiatechs/jsonata-go"
)

func NewCompiler() language.Compiler {
	return language.CompileFunc(func(code string) (language.Program, error) {
		exp, err := jsonata.Compile(code)
		if err != nil {
			return nil, err
		}
		return language.RunFunc(func(env any) (any, error) {
			return exp.Eval(env)
		}), nil
	})
}
