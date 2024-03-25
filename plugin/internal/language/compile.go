package language

import (
	"encoding/json"
	"fmt"
	"github.com/dop251/goja"
	"github.com/evanw/esbuild/pkg/api"
	"github.com/pkg/errors"
	"github.com/samber/lo"
	"github.com/siyul-park/uniflow/pkg/primitive"
	"github.com/siyul-park/uniflow/plugin/internal/js"
	"github.com/xiatechs/jsonata-go"
	"gopkg.in/yaml.v3"
	"sync"
)

func CompileTransformWithPrimitive(code string, lang string) (func(primitive.Value) (primitive.Value, error), error) {
	transform, err := CompileTransform(code, &lang)
	if err != nil {
		return nil, err
	}

	return func(value primitive.Value) (primitive.Value, error) {
		var input any
		switch lang {
		case Typescript, Javascript, JSONata:
			input = primitive.Interface(value)
		}
		if output, err := transform(input); err != nil {
			return nil, err
		} else {
			return primitive.MarshalBinary(output)
		}
	}, nil
}

func CompileTransform(code string, lang *string) (func(any) (any, error), error) {
	if lang == nil {
		lang = lo.ToPtr("")
	}
	if *lang == "" {
		*lang = Detect(code)
	}

	switch *lang {
	case Text, JSON, YAML:
		var data any
		var err error
		if *lang == Text {
			data = code
		} else if *lang == JSON {
			err = json.Unmarshal([]byte(code), &data)
		} else if *lang == YAML {
			err = yaml.Unmarshal([]byte(code), &data)
		}
		if err != nil {
			return nil, err
		}

		return func(_ any) (any, error) {
			return data, nil
		}, nil
	case Javascript, Typescript:
		if !js.AssertExportFunction(code, "default") {
			code = fmt.Sprintf("module.exports = ($) => { return (%s); }", code)
		}

		var err error
		if *lang == Typescript {
			if code, err = js.Transform(code, api.TransformOptions{Loader: api.LoaderTS}); err != nil {
				return nil, err
			}
		}
		if code, err = js.Transform(code, api.TransformOptions{Format: api.FormatCommonJS}); err != nil {
			return nil, err
		}

		program, err := goja.Compile("", code, true)
		if err != nil {
			return nil, err
		}

		vms := &sync.Pool{
			New: func() any {
				vm := js.New()
				_, _ = vm.RunProgram(program)
				return vm
			},
		}

		return func(input any) (any, error) {
			vm := vms.Get().(*goja.Runtime)
			defer vms.Put(vm)

			defaults := js.Export(vm, "default")
			argument, _ := goja.AssertFunction(defaults)

			if output, err := argument(goja.Undefined(), vm.ToValue(input)); err != nil {
				return false, err
			} else {
				return output.Export(), nil
			}
		}, nil
	case JSONata:
		exp, err := jsonata.Compile(code)
		if err != nil {
			return nil, err
		}
		return func(input any) (any, error) {
			if output, err := exp.Eval(input); err != nil {
				return false, err
			} else {
				return output, nil
			}
		}, nil
	default:
		return nil, errors.WithStack(ErrUnsupportedLanguage)
	}
}
