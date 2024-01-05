package control

import (
	"encoding/json"

	"github.com/pkg/errors"
	"github.com/siyul-park/uniflow/pkg/node"
	"github.com/siyul-park/uniflow/pkg/packet"
	"github.com/siyul-park/uniflow/pkg/primitive"
	"github.com/siyul-park/uniflow/pkg/process"
	"github.com/siyul-park/uniflow/pkg/scheme"
	"github.com/xiatechs/jsonata-go"
	"gopkg.in/yaml.v3"
)

type SnippetNode struct {
	*node.OneToOneNode
}

const (
	LangJSON    = "json"
	LangYAML    = "yaml"
	LangJSONata = "jsonata"
)

var ErrInvalidLanguage = errors.New("language is invalid")

var _ node.Node = (*SnippetNode)(nil)

func NewSnippetNode(lang, code string) (*SnippetNode, error) {
	action, err := compile(lang, code)
	if err != nil {
		return nil, err
	}
	return &SnippetNode{OneToOneNode: node.NewOneToOneNode(action)}, nil
}

func compile(lang, code string) (func(*process.Process, *packet.Packet) (*packet.Packet, *packet.Packet), error) {
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
	default:
		return nil, errors.WithStack(ErrInvalidLanguage)
	}
}

type SnippetNodeSpec struct {
	scheme.SpecMeta
	Lang string `map:"lang"`
	Code string `map:"code"`
}

func NewSnippetNodeCodec() scheme.Codec {
	return scheme.CodecWithType[*SnippetNodeSpec](func(spec *SnippetNodeSpec) (node.Node, error) {
		return NewSnippetNode(spec.Lang, spec.Code)
	})
}
