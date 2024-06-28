package javascript

import (
	"errors"
	"reflect"
	"strings"
	"sync"

	"github.com/dop251/goja"
	"github.com/evanw/esbuild/pkg/api"
	"github.com/iancoleman/strcase"
	"github.com/siyul-park/uniflow/ext/language"
)

type fieldNameMapper struct{}

const (
	keyModule  = "module"
	keyExports = "exports"
	keyDefault = "default"
)

var _ goja.FieldNameMapper = &fieldNameMapper{}

func NewCompiler(options ...api.TransformOptions) language.Compiler {
	if len(options) == 0 {
		options = append(options, api.TransformOptions{
			Format: api.FormatCommonJS,
		})
	}

	return language.CompileFunc(func(code string) (language.Program, error) {
		for _, options := range options {
			result := api.Transform(code, options)
			if len(result.Errors) > 0 {
				var msgs []string
				for _, err := range result.Errors {
					msgs = append(msgs, err.Text)
				}
				return nil, errors.New(strings.Join(msgs, ", "))
			}
			code = string(result.Code)
		}

		program, err := goja.Compile("", code, true)
		if err != nil {
			return nil, err
		}

		vms := sync.Pool{
			New: func() any {
				vm := goja.New()
				vm.SetFieldNameMapper(&fieldNameMapper{})
				module := vm.NewObject()
				exports := vm.NewObject()

				_ = module.Set(keyExports, exports)

				_ = vm.Set(keyModule, module)
				_ = vm.Set(keyExports, exports)

				vm.RunProgram(program)
				return vm
			},
		}

		return language.RunFunc(func(env any) (any, error) {
			vm := vms.Get().(*goja.Runtime)
			defer vms.Put(vm)

			module := vm.Get(keyModule)
			if module == nil {
				return nil, nil
			}

			exports := module.ToObject(vm).Get(keyExports)
			if exports == nil {
				return nil, nil
			}

			exModule := exports.ToObject(vm).Get("__esModule")
			if exModule != nil && exModule.Export() == true {
				exports = exports.ToObject(vm).Get(keyDefault)
			}

			run, ok := goja.AssertFunction(exports)
			if !ok {
				return nil, nil
			}

			if result, err := run(goja.Undefined(), vm.ToValue(env)); err != nil {
				return nil, err
			} else {
				return result.Export(), nil
			}

		}), nil
	})
}

func (*fieldNameMapper) FieldName(_ reflect.Type, f reflect.StructField) string {
	return strcase.ToLowerCamel(f.Name)
}

func (*fieldNameMapper) MethodName(_ reflect.Type, m reflect.Method) string {
	return strcase.ToLowerCamel(m.Name)
}
