package control

import (
	"encoding/json"
	"sync"

	"github.com/dop251/goja"
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
	LangJSON       = "json"
	LangYAML       = "yaml"
	LangJavascript = "javascript"
	LangJSONata    = "jsonata"
)

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
	n := &SnippetNode{}
	action, err := n.compile(lang, code)
	if err != nil {
		return nil, err
	}
	n.OneToOneNode = node.NewOneToOneNode(action)
	return n, nil
}

func (n *SnippetNode) compile(lang, code string) (func(*process.Process, *packet.Packet) (*packet.Packet, *packet.Packet), error) {
	switch lang {
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
	case LangJavascript:
		program, err := goja.Compile("", code, true)
		if err != nil {
			return nil, err
		}

		vm := goja.New()
		_, err = vm.RunProgram(program)
		if err != nil {
			return nil, err
		}

		_, ok := goja.AssertFunction(vm.Get("main"))
		if !ok {
			return nil, errors.WithStack(ErrEntryPointNotUndeclared)
		}

		vmPool := &sync.Pool{
			New: func() any {
				vm := goja.New()
				_, _ = vm.RunProgram(program)
				return vm
			},
		}

		return func(proc *process.Process, inPck *packet.Packet) (*packet.Packet, *packet.Packet) {
			vm := vmPool.Get().(*goja.Runtime)
			defer vmPool.Put(vm)

			main, ok := goja.AssertFunction(vm.Get("main"))
			if !ok {
				return nil, packet.WithError(ErrEntryPointNotUndeclared, inPck)
			}

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

	outPayload := primitive.NewString(code)

	return func(proc *process.Process, _ *packet.Packet) (*packet.Packet, *packet.Packet) {
		return packet.New(outPayload), nil
	}, nil
}
