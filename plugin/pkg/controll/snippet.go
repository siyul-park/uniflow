package controll

import (
	"encoding/json"

	"github.com/pkg/errors"
	"github.com/siyul-park/uniflow/pkg/node"
	"github.com/siyul-park/uniflow/pkg/packet"
	"github.com/siyul-park/uniflow/pkg/primitive"
	"github.com/siyul-park/uniflow/pkg/process"
)

type SnippetNode struct {
	*node.OneToOneNode
}

const (
	LangJSON = "json"
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

func compile(lang, code string) (func(proc *process.Process, inPck *packet.Packet) (*packet.Packet, *packet.Packet), error) {
	switch lang {
	case LangJSON:
		var data any
		err := json.Unmarshal([]byte(code), &data)
		if err != nil {
			return nil, err
		}
		outPayload, err := primitive.MarshalBinary(data)
		if err != nil {
			return nil, err
		}
		return func(proc *process.Process, inPck *packet.Packet) (*packet.Packet, *packet.Packet) {
			return packet.New(outPayload), nil
		}, nil
	default:
		return nil, errors.WithStack(ErrInvalidLanguage)
	}
}
