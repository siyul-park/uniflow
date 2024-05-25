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
	query func(any) (string, error)
	args  func(any) (any, error)
	mu    sync.RWMutex
}

// SQLNodeSpec holds the specifications for creating a SQLNode.
type SQLNodeSpec struct {
	scheme.SpecMeta `map:",inline"`
	Lang            string `map:"lang,omitempty"`
	Query           string `map:"query"`
	Arguments       string `map:"arguments,omitempty"`
}

const KindSQL = "sql"

// NewSQLNode creates a new SQLNode instance.
func NewSQLNode(query, lang string) (*SQLNode, error) {
	l := lang
	transform, err := language.CompileTransform(query, &l)
	if err != nil {
		return nil, err
	}

	n := &SQLNode{lang: lang}
	n.OneToOneNode = node.NewOneToOneNode(n.action)
	n.query = func(input any) (string, error) {
		output, err := transform(input)
		if err != nil {
			return "", err
		}
		if output, ok := output.(string); !ok {
			return "", errors.WithStack(packet.ErrInvalidPacket)
		} else {
			return output, nil
		}
	}

	return n, nil
}

// SetArguments sets the arguments for the SQL query.
func (n *SQLNode) SetArguments(args string) error {
	n.mu.Lock()
	defer n.mu.Unlock()

	if args == "" {
		n.args = nil
		return nil
	}

	lang := n.lang
	transform, err := language.CompileTransform(args, &lang)
	if err != nil {
		return err
	}

	n.args = transform

	return nil
}

func (n *SQLNode) action(proc *process.Process, inPck *packet.Packet) (*packet.Packet, *packet.Packet) {
	n.mu.RLock()
	defer n.mu.RUnlock()

	inPayload := inPck.Payload()
	input := primitive.Interface(inPayload)

	query, err := n.query(input)
	if err != nil {
		return nil, packet.WithError(err)
	}

	if n.args == nil {
		outPayload, err := primitive.MarshalText(query)
		if err != nil {
			return nil, packet.WithError(err)
		}
		return packet.New(outPayload), nil
	}

	args, err := n.args(input)
	if err != nil {
		return nil, packet.WithError(err)
	}

	outPayload, err := primitive.MarshalText([]any{query, args})
	if err != nil {
		return nil, packet.WithError(err)
	}
	return packet.New(outPayload), nil
}

// NewSQLNodeCodec creates a new codec for SQLNodeSpec.
func NewSQLNodeCodec() scheme.Codec {
	return scheme.CodecWithType(func(spec *SQLNodeSpec) (node.Node, error) {
		n, err := NewSQLNode(spec.Query, spec.Lang)
		if err != nil {
			return nil, err
		}
		if err := n.SetArguments(spec.Arguments); err != nil {
			_ = n.Close()
			return nil, err
		}
		return n, nil
	})
}
