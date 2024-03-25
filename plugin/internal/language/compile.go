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

func CompileTransformWithPrimitive(template primitive.Value, lang string) (func(primitive.Value) (primitive.Value, error), error) {
	switch v := template.(type) {
	case *primitive.Map:
		transforms := make([]func(primitive.Value) (primitive.Value, error), 0, v.Len())
		for _, k := range v.Keys() {
			if transform, err := CompileTransformWithPrimitive(v.GetOr(k, nil), lang); err != nil {
				return nil, err
			} else {
				transforms = append(transforms, transform)
			}
		}

		return func(value primitive.Value) (primitive.Value, error) {
			pairs := make([]primitive.Value, 0, v.Len()*2)
			for i, k := range v.Keys() {
				transform := transforms[i]
				if v, err := transform(value); err != nil {
					return nil, err
				} else {
					pairs = append(pairs, k)
					pairs = append(pairs, v)
				}
			}
			return primitive.NewMap(pairs...), nil
		}, nil
	case *primitive.Slice:
		transforms := make([]func(primitive.Value) (primitive.Value, error), 0, v.Len())
		for _, v := range v.Values() {
			if transform, err := CompileTransformWithPrimitive(v, lang); err != nil {
				return nil, err
			} else {
				transforms = append(transforms, transform)
			}
		}

		return func(value primitive.Value) (primitive.Value, error) {
			values := make([]primitive.Value, 0, v.Len()*2)
			for i, v := range v.Values() {
				transform := transforms[i]
				if v, err := transform(v); err != nil {
					return nil, err
				} else {
					values = append(values, v)
				}
			}
			return primitive.NewSlice(values...), nil
		}, nil
	case primitive.String:
		transform, err := CompileTransform(v.String(), &lang)
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
	default:
		return func(value primitive.Value) (primitive.Value, error) {
			return v, nil
		}, nil
	}
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
