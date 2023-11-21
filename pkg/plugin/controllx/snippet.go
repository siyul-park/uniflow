package controllx

import (
	"encoding/json"
	"reflect"
	"strings"
	"sync"

	"github.com/dop251/goja"
	"github.com/evanw/esbuild/pkg/api"
	"github.com/iancoleman/strcase"
	"github.com/oklog/ulid/v2"
	"github.com/pkg/errors"
	"github.com/siyul-park/uniflow/pkg/node"
	"github.com/siyul-park/uniflow/pkg/packet"
	"github.com/siyul-park/uniflow/pkg/primitive"
	"github.com/siyul-park/uniflow/pkg/process"
	"github.com/siyul-park/uniflow/pkg/scheme"
	"github.com/xiatechs/jsonata-go"
)

type (
	SnippetNodeConfig struct {
		ID   ulid.ULID
		Lang string
		Code string
	}

	SnippetNode struct {
		*node.OneToOneNode
		run func(any) (any, error)
	}

	SnippetSpec struct {
		scheme.SpecMeta `map:",inline"`
		Lang            string `map:"lang"`
		Code            string `map:"code"`
	}

	fieldNameMapper struct{}
)

const (
	KindSnippet = "snippet"
)

const (
	LangTypescript = "typescript"
	LangJavascript = "javascript"
	LangJSON       = "json"
	LangJSONata    = "jsonata"
)

var _ node.Node = &SnippetNode{}

var (
	ErrEntryPointNotUndeclared = errors.New("entry point is undeclared")
	ErrNotSupportedLanguage    = errors.New("language is not supported")
)

func NewSnippetNode(config SnippetNodeConfig) (*SnippetNode, error) {
	defer func() { _ = recover() }()

	id := config.ID
	lang := config.Lang
	code := config.Code

	run, err := compile(lang, code)
	if err != nil {
		return nil, err
	}

	n := &SnippetNode{
		run: run,
	}
	n.OneToOneNode = node.NewOneToOneNode(node.OneToOneNodeConfig{
		ID:     id,
		Action: n.action,
	})

	return n, nil
}

func (n *SnippetNode) action(proc *process.Process, inPck *packet.Packet) (*packet.Packet, *packet.Packet) {
	inPayload := inPck.Payload()

	var input any
	if inPayload != nil {
		input = inPayload.Interface()
	}

	if output, err := n.run(input); err != nil {
		return nil, packet.NewError(err, inPck)
	} else if outPayload, err := primitive.MarshalText(output); err != nil {
		return nil, packet.NewError(err, inPck)
	} else {
		return packet.New(outPayload), nil
	}
}

func compile(lang, code string) (func(any) (any, error), error) {
	switch lang {
	case LangJSON:
		var val any
		if err := json.Unmarshal([]byte(code), &val); err != nil {
			return nil, err
		}

		return func(payload any) (any, error) {
			return val, nil
		}, nil
	case LangTypescript, LangJavascript:
		if lang == LangTypescript {
			result := api.Transform(code, api.TransformOptions{
				Loader: api.LoaderTS,
			})
			if len(result.Errors) > 0 {
				var msgs []string
				for _, msg := range result.Errors {
					msgs = append(msgs, msg.Text)
				}
				return nil, errors.New(strings.Join(msgs, ", "))
			}
			code = string(result.Code)
		}
		program, err := goja.Compile("", code, true)
		if err != nil {
			return nil, err
		}

		vm := goja.New()
		if _, err := vm.RunProgram(program); err != nil {
			return nil, err
		}
		if _, ok := goja.AssertFunction(vm.Get("main")); !ok {
			return nil, errors.WithStack(ErrEntryPointNotUndeclared)
		}

		vmPool := &sync.Pool{
			New: func() any {
				vm := goja.New()
				_, _ = vm.RunProgram(program)
				vm.SetFieldNameMapper(&fieldNameMapper{})
				return vm
			},
		}

		return func(payload any) (any, error) {
			vm := vmPool.Get().(*goja.Runtime)
			defer vmPool.Put(vm)

			main, ok := goja.AssertFunction(vm.Get("main"))
			if !ok {
				return nil, errors.WithStack(ErrEntryPointNotUndeclared)
			}

			if output, err := main(goja.Undefined(), vm.ToValue(payload)); err != nil {
				return nil, err
			} else {
				return output.Export(), nil
			}
		}, nil
	case LangJSONata:
		exp, err := jsonata.Compile(code)
		if err != nil {
			return nil, err
		}

		return func(payload any) (any, error) {
			if output, err := exp.Eval(payload); err != nil {
				return nil, err
			} else {
				return output, nil
			}
		}, nil
	default:
		return nil, errors.WithStack(ErrNotSupportedLanguage)
	}
}

func (*fieldNameMapper) FieldName(t reflect.Type, f reflect.StructField) string {
	return strcase.ToLowerCamel(f.Name)
}

func (*fieldNameMapper) MethodName(t reflect.Type, m reflect.Method) string {
	return strcase.ToLowerCamel(m.Name)
}