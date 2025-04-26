package javascript

import (
	"context"
	"errors"
	"reflect"
	"strings"
	"sync"

	"github.com/dop251/goja"
	"github.com/evanw/esbuild/pkg/api"
	"github.com/iancoleman/strcase"
	"github.com/siyul-park/uniflow/pkg/language"
)

type fieldNameMapper struct{}

const Language = "javascript"

const (
	keyModule   = "module"
	keyExports  = "exports"
	keyDefault  = "default"
	keyESModule = "__esModule"
)

var _ goja.FieldNameMapper = &fieldNameMapper{}

func NewCompiler(options ...api.TransformOptions) language.Compiler {
	if len(options) == 0 {
		options = append(options, api.TransformOptions{
			Format: api.FormatCommonJS,
			Target: api.ES2016,
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

				_, _ = vm.RunProgram(program)
				return vm
			},
		}

		return language.RunFunc(func(ctx context.Context, args ...any) (_ any, err error) {
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

			exModule := exports.ToObject(vm).Get(keyESModule)
			if exModule != nil && exModule.Export() == true {
				exports = exports.ToObject(vm).Get(keyDefault)
			}

			run, ok := goja.AssertFunction(exports)
			if !ok {
				return nil, nil
			}

			values := make([]goja.Value, 0, len(args))
			for _, arg := range args {
				values = append(values, vm.ToValue(arg))
			}

			done := make(chan struct{})
			defer close(done)

			go func() {
				select {
				case <-ctx.Done():
					vm.Interrupt(ctx.Err())
				case <-done:
				}
			}()

			if result, err := run(vm.ToValue(ctx), values...); err != nil {
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
