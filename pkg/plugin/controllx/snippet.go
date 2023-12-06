package controllx

import (
	"encoding/json"
	"reflect"
	"strings"
	"sync"

	"github.com/dop251/goja"
	"github.com/evanw/esbuild/pkg/api"
	"github.com/iancoleman/strcase"
	"github.com/pkg/errors"
	"github.com/siyul-park/uniflow/pkg/node"
	"github.com/siyul-park/uniflow/pkg/packet"
	"github.com/siyul-park/uniflow/pkg/primitive"
	"github.com/siyul-park/uniflow/pkg/process"
	"github.com/siyul-park/uniflow/pkg/scheme"
	"github.com/xiatechs/jsonata-go"
)

// SnippetNodeConfig holds the configuration for creating a SnippetNode.
type SnippetNodeConfig struct {
	Lang string
	Code string
}

// SnippetNode represents a node that runs a snippet of code.
type SnippetNode struct {
	*node.OneToOneNode
	run func(any) (any, error)
}

// SnippetSpec represents the specification for the SnippetNode.
type SnippetSpec struct {
	scheme.SpecMeta `map:",inline"`
	Lang            string `map:"lang"`
	Code            string `map:"code"`
}

// KindSnippet is the kind identifier for SnippetNode.
const KindSnippet = "snippet"

// Supported programming languages for snippets.
const (
	LangTypescript = "typescript"
	LangJavascript = "javascript"
	LangJSON       = "json"
	LangJSONata    = "jsonata"
)

// Errors related to snippet execution.
var (
	ErrEntryPointNotUndeclared = errors.New("entry point is undeclared")
	ErrNotSupportedLanguage    = errors.New("language is not supported")
)

var _ node.Node = (*SnippetNode)(nil)
var _ scheme.Spec = (*SnippetSpec)(nil)

// NewSnippetNode creates a new SnippetNode with the given configuration.
func NewSnippetNode(config SnippetNodeConfig) (*SnippetNode, error) {
	defer func() { _ = recover() }()

	lang := config.Lang
	code := config.Code

	run, err := compile(lang, code)
	if err != nil {
		return nil, err
	}

	n := &SnippetNode{
		run: run,
	}
	n.OneToOneNode = node.NewOneToOneNode(n.action)

	return n, nil
}

func (n *SnippetNode) action(proc *process.Process, inPck *packet.Packet) (*packet.Packet, *packet.Packet) {
	inPayload := inPck.Payload()

	var input any
	if inPayload != nil {
		input = inPayload.Interface()
	}

	if output, err := n.run(input); err != nil {
		return nil, packet.WithError(err, inPck)
	} else if outPayload, err := primitive.MarshalText(output); err != nil {
		return nil, packet.WithError(err, inPck)
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

type fieldNameMapper struct{}

func (*fieldNameMapper) FieldName(t reflect.Type, f reflect.StructField) string {
	return strcase.ToLowerCamel(f.Name)
}

func (*fieldNameMapper) MethodName(t reflect.Type, m reflect.Method) string {
	return strcase.ToLowerCamel(m.Name)
}
