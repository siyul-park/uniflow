package datastore

import (
	"sync"

	"github.com/pkg/errors"
	"github.com/siyul-park/uniflow/pkg/node"
	"github.com/siyul-park/uniflow/pkg/packet"
	"github.com/siyul-park/uniflow/pkg/primitive"
	"github.com/siyul-park/uniflow/pkg/process"
	"github.com/siyul-park/uniflow/pkg/scheme"
	"github.com/siyul-park/uniflow/plugin/internal/language"
)

// SQLNode represents a node for executing SQL queries.
type SQLNode struct {
	*node.OneToOneNode
	lang  string
	query func(primitive.Value) (primitive.String, error)
	args  func(primitive.Value) (primitive.Value, error)
	mu    sync.RWMutex
}

// SQLNodeSpec holds the specifications for creating a SQLNode.
type SQLNodeSpec struct {
	scheme.SpecMeta `map:",inline"`
	Lang            string `map:"lang,omitempty"`
	Query           string `map:"query"`
	Args            string `map:"args,omitempty"`
}

const KindSQL = "sql"

func NewSQLNode(query, lang string) (*SQLNode, error) {
	transform, err := language.CompileTransformWithPrimitive(query, lang)
	if err != nil {
		return nil, err
	}

	n := &SQLNode{lang: lang}
	n.OneToOneNode = node.NewOneToOneNode(n.action)
	n.query = func(value primitive.Value) (primitive.String, error) {
		output, err := transform(value)
		if err != nil {
			return primitive.String(""), err
		}
		if output, ok := output.(primitive.String); !ok {
			return primitive.String(""), errors.WithStack(packet.ErrInvalidPacket)
		} else {
			return output, nil
		}
	}

	return n, nil
}

func (n *SQLNode) SetArguments(args string) error {
	n.mu.Lock()
	defer n.mu.Unlock()

	if args == "" {
		n.args = nil
		return nil
	}

	arguments, err := language.CompileTransformWithPrimitive(args, n.lang)
	if err != nil {
		return err
	}
	n.args = arguments

	return nil
}

func (n *SQLNode) action(proc *process.Process, inPck *packet.Packet) (*packet.Packet, *packet.Packet) {
	n.mu.RLock()
	defer n.mu.RUnlock()

	inPayload := inPck.Payload()

	query, err := n.query(inPayload)
	if err != nil {
		return nil, packet.WithError(err, inPck)
	}

	if n.args == nil {
		return packet.New(query), nil
	}

	args, err := n.args(inPayload)
	if err != nil {
		return nil, packet.WithError(err, inPck)
	}

	outPayload := primitive.NewSlice(query, args)
	return packet.New(outPayload), nil
}

// NewSQLNodeCodec creates a new codec for SQLNodeSpec.
func NewSQLNodeCodec() scheme.Codec {
	return scheme.CodecWithType(func(spec *SQLNodeSpec) (node.Node, error) {
		n, err := NewSQLNode(spec.Query, spec.Lang)
		if err != nil {
			return nil, err
		}
		if err := n.SetArguments(spec.Args); err != nil {
			_ = n.Close()
			return nil, err
		}
		return n, nil
	})
}
