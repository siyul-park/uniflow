package control

import (
	"encoding/json"
	"strings"
	"sync"

	"github.com/dop251/goja"
	"github.com/evanw/esbuild/pkg/api"
	"github.com/pkg/errors"
	"github.com/siyul-park/uniflow/pkg/node"
	"github.com/siyul-park/uniflow/pkg/packet"
	"github.com/siyul-park/uniflow/pkg/primitive"
	"github.com/siyul-park/uniflow/pkg/process"
	"github.com/siyul-park/uniflow/pkg/scheme"
	"github.com/xiatechs/jsonata-go"
	"gopkg.in/yaml.v3"
)

// SnippetNode represents a node that executes code snippets in various languages.
type SnippetNode struct {
	*node.OneToOneNode
}

// SnippetNodeSpec holds the specifications for creating a SnippetNode.
type SnippetNodeSpec struct {
	scheme.SpecMeta

	Lang string `map:"lang"`
	Code string `map:"code"`
}

const KindSnippet = "snippet"

const (
	LangText       = "text"
	LangTypescript = "typescript"
	LangJSON       = "json"
	LangYAML       = "yaml"
	LangJavascript = "javascript"
	LangJSONata    = "jsonata"
)

var _ node.Node = (*SnippetNode)(nil)

var (
	ErrUnsupportedLanguage     = errors.New("language not supported")
	ErrEntryPointNotUndeclared = errors.New("entry point not defined")
)

// NewSnippetNodeCodec creates a new codec for SnippetNodeSpec.
func NewSnippetNodeCodec() scheme.Codec {
	return scheme.CodecWithType[*SnippetNodeSpec](func(spec *SnippetNodeSpec) (node.Node, error) {
		return NewSnippetNode(spec.Lang, spec.Code)
	})
}

// NewSnippetNode creates a new SnippetNode with the specified language and code.
func NewSnippetNode(lang, code string) (*SnippetNode, error) {
	n := &SnippetNode{}
	action, err := n.compile(lang, code)
	if err != nil {
		return nil, err
	}
	n.OneToOneNode = node.NewOneToOneNode(action)
	return n, nil
}

func (n *SnippetNode) compile(lang, code string) (func(*process.Process, *packet.Packet) (*packet.Packet, *packet.Packet), error) {
	if lang == "" {
		lang = LangText
	}

	switch lang {
	case LangText:
		outPayload := primitive.NewString(code)

		return func(proc *process.Process, _ *packet.Packet) (*packet.Packet, *packet.Packet) {
			return packet.New(outPayload), nil
		}, nil
	case LangJSON, LangYAML:
		var data any
		var err error
		if lang == LangJSON {
			err = json.Unmarshal([]byte(code), &data)
		} else if lang == LangYAML {
			err = yaml.Unmarshal([]byte(code), &data)
		}
		if err != nil {
			return nil, err
		}

		outPayload, err := primitive.MarshalBinary(data)
		if err != nil {
			return nil, err
		}

		return func(proc *process.Process, _ *packet.Packet) (*packet.Packet, *packet.Packet) {
			return packet.New(outPayload), nil
		}, nil
	case LangJavascript, LangTypescript:
		if lang == LangTypescript {
			var err error
			if code, err = transformJavascript(code, api.TransformOptions{Loader: api.LoaderTS}); err != nil {
				return nil, err
			}
		}
		var err error
		if code, err = transformJavascript(code, api.TransformOptions{Format: api.FormatCommonJS}); err != nil {
			return nil, err
		}

		program, err := goja.Compile("", code, true)
		if err != nil {
			return nil, err
		}

		vm := goja.New()
		if err := useCommonJSModule(vm); err != nil {
			return nil, err
		}
		_, err = vm.RunProgram(program)
		if err != nil {
			return nil, err
		}

		defaults := getCommonJSExport(vm, "default")
		if defaults == nil {
			return nil, errors.WithStack(ErrEntryPointNotUndeclared)
		}
		_, ok := goja.AssertFunction(defaults)
		if !ok {
			return nil, errors.WithStack(ErrEntryPointNotUndeclared)
		}

		vmPool := &sync.Pool{
			New: func() any {
				vm := goja.New()
				_ = useCommonJSModule(vm)
				_, _ = vm.RunProgram(program)
				return vm
			},
		}

		return func(proc *process.Process, inPck *packet.Packet) (*packet.Packet, *packet.Packet) {
			vm := vmPool.Get().(*goja.Runtime)
			defer vmPool.Put(vm)

			defaults := getCommonJSExport(vm, "default")
			main, _ := goja.AssertFunction(defaults)

			inPayload := inPck.Payload()
			input := inPayload.Interface()

			if output, err := main(goja.Undefined(), vm.ToValue(input)); err != nil {
				return nil, packet.WithError(err, inPck)
			} else if outPayload, err := primitive.MarshalBinary(output.Export()); err != nil {
				return nil, packet.WithError(err, inPck)
			} else {
				return packet.New(outPayload), nil
			}
		}, nil
	case LangJSONata:
		exp, err := jsonata.Compile(code)
		if err != nil {
			return nil, err
		}

		return func(proc *process.Process, inPck *packet.Packet) (*packet.Packet, *packet.Packet) {
			inPayload := inPck.Payload()
			input := inPayload.Interface()

			output, err := exp.Eval(input)
			if err != nil {
				return nil, packet.WithError(err, inPck)
			}
			outPayload, err := primitive.MarshalBinary(output)
			if err != nil {
				return nil, packet.WithError(err, inPck)
			}

			return packet.New(outPayload), nil
		}, nil
	}

	return nil, ErrUnsupportedLanguage
}

func transformJavascript(code string, options api.TransformOptions) (string, error) {
	if result := api.Transform(code, options); len(result.Errors) > 0 {
		var msgs []string
		for _, msg := range result.Errors {
			msgs = append(msgs, msg.Text)
		}
		return "", errors.New(strings.Join(msgs, ", "))
	} else {
		return string(result.Code), nil
	}
}

func useCommonJSModule(vm *goja.Runtime) error {
	module := vm.NewObject()
	exports := vm.NewObject()

	if err := vm.Set("module", module); err != nil {
		return err
	}
	if err := vm.Set("exports", exports); err != nil {
		return err
	}
	if err := module.Set("exports", exports); err != nil {
		return err
	}
	return nil
}

func getCommonJSExport(vm *goja.Runtime, name string) goja.Value {
	module := vm.Get("module")
	if module == nil {
		return nil
	}
	exports := module.ToObject(vm).Get("exports")
	if exports == nil {
		return nil
	}

	if name == "default" {
		if exports.ToObject(vm).Get("__esModule").Export() == true {
			return exports.ToObject(vm).Get("default")
		} else {
			return exports
		}
	}
	return exports.ToObject(vm).Get(name)
}
