package control

import (
	"encoding/json"
	"sync"

	"github.com/dop251/goja"
	"github.com/evanw/esbuild/pkg/api"
	"github.com/pkg/errors"
	"github.com/siyul-park/uniflow/pkg/node"
	"github.com/siyul-park/uniflow/pkg/packet"
	"github.com/siyul-park/uniflow/pkg/primitive"
	"github.com/siyul-park/uniflow/pkg/process"
	"github.com/siyul-park/uniflow/pkg/scheme"
	"github.com/siyul-park/uniflow/plugin/internal/js"
	"github.com/xiatechs/jsonata-go"
	"gopkg.in/yaml.v3"
)

// SnippetNode represents a node that executes code snippets in various languages.
type SnippetNode struct {
	*node.OneToOneNode
	lang   string
	code   string
	action func(*process.Process, *packet.Packet) (*packet.Packet, *packet.Packet)
	mu     sync.RWMutex
}

// SnippetNodeSpec holds the specifications for creating a SnippetNode.
type SnippetNodeSpec struct {
	scheme.SpecMeta
	Lang string `map:"lang"`
	Code string `map:"code"`
}

const KindSnippet = "snippet"

var _ node.Node = (*SnippetNode)(nil)

var ErrEntryPointNotUndeclared = errors.New("entry point not defined")

// NewSnippetNodeCodec creates a new codec for SnippetNodeSpec.
func NewSnippetNodeCodec() scheme.Codec {
	return scheme.CodecWithType[*SnippetNodeSpec](func(spec *SnippetNodeSpec) (node.Node, error) {
		return NewSnippetNode(spec.Lang, spec.Code)
	})
}

// NewSnippetNode creates a new SnippetNode with the specified language and code.
func NewSnippetNode(lang, code string) (*SnippetNode, error) {
	n := &SnippetNode{lang: lang, code: code}
	if err := n.compile(); err != nil {
		return nil, err
	}
	if err := n.link(); err != nil {
		return nil, err
	}
	return n, nil
}

func (n *SnippetNode) compile() error {
	n.mu.Lock()
	defer n.mu.Unlock()

	if n.lang == "" {
		n.lang = LangText
	}

	switch n.lang {
	case LangText:
		outPayload := primitive.NewString(n.code)
		n.action = func(proc *process.Process, _ *packet.Packet) (*packet.Packet, *packet.Packet) {
			return packet.New(outPayload), nil
		}
		return nil
	case LangJSON, LangYAML:
		var data any
		var err error
		if n.lang == LangJSON {
			err = json.Unmarshal([]byte(n.code), &data)
		} else if n.lang == LangYAML {
			err = yaml.Unmarshal([]byte(n.code), &data)
		}
		if err != nil {
			return err
		}

		outPayload, err := primitive.MarshalBinary(data)
		if err != nil {
			return err
		}

		n.action = func(proc *process.Process, _ *packet.Packet) (*packet.Packet, *packet.Packet) {
			return packet.New(outPayload), nil
		}
		return nil
	case LangJavascript, LangTypescript:
		var err error
		if n.lang == LangTypescript {
			if n.code, err = js.Transform(n.code, api.TransformOptions{Loader: api.LoaderTS}); err != nil {
				return err
			}
		}
		if n.code, err = js.Transform(n.code, api.TransformOptions{Format: api.FormatCommonJS}); err != nil {
			return err
		}

		program, err := goja.Compile("", n.code, true)
		if err != nil {
			return err
		}

		vm := js.New()
		if _, err := vm.RunProgram(program); err != nil {
			return err
		}

		if defaults := js.Export(vm, "default"); defaults == nil {
			return errors.WithStack(ErrEntryPointNotUndeclared)
		} else if _, ok := goja.AssertFunction(defaults); !ok {
			return errors.WithStack(ErrEntryPointNotUndeclared)
		}

		vms := &sync.Pool{
			New: func() any {
				vm := js.New()
				_, _ = vm.RunProgram(program)
				return vm
			},
		}

		n.action = func(proc *process.Process, inPck *packet.Packet) (*packet.Packet, *packet.Packet) {
			vm := vms.Get().(*goja.Runtime)
			defer vms.Put(vm)

			defaults := js.Export(vm, "default")
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
		}
		return nil
	case LangJSONata:
		exp, err := jsonata.Compile(n.code)
		if err != nil {
			return err
		}

		n.action = func(proc *process.Process, inPck *packet.Packet) (*packet.Packet, *packet.Packet) {
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
		}
		return nil
	}

	return ErrUnsupportedLanguage
}

func (n *SnippetNode) link() error {
	n.mu.Lock()
	defer n.mu.Unlock()

	if n.OneToOneNode != nil {
		if err := n.OneToOneNode.Close(); err != nil {
			return err
		}
	}

	n.OneToOneNode = node.NewOneToOneNode(n.action)
	return nil
}
