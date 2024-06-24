package javascript

import (
	"errors"
	"strings"

	"github.com/dop251/goja"
	"github.com/evanw/esbuild/pkg/api"
	"github.com/siyul-park/uniflow/plugin/pkg/language"
)

type compiler struct {
	options api.TransformOptions
}

var _ language.Compiler = (*compiler)(nil)

func NewCompiler(options api.TransformOptions) language.Compiler {
	if options.Format == 0 {
		options.Format = api.FormatCommonJS
	}
	return &compiler{
		options: options,
	}
}

func (c *compiler) Compile(code string) (language.Program, error) {
	result := api.Transform(code, c.options)
	if len(result.Errors) > 0 {
		var msgs []string
		for _, err := range result.Errors {
			msgs = append(msgs, err.Text)
		}
		return nil, errors.New(strings.Join(msgs, ", "))
	}

	program, err := goja.Compile("", string(result.Code), true)
	if err != nil {
		return nil, err
	}
	return newProgram(program)
}
