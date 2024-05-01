package datastore

import (
	"database/sql"
	"sync"

	"github.com/jmoiron/sqlx"
	"github.com/siyul-park/uniflow/pkg/node"
	"github.com/siyul-park/uniflow/pkg/packet"
	"github.com/siyul-park/uniflow/pkg/primitive"
	"github.com/siyul-park/uniflow/pkg/process"
	"github.com/siyul-park/uniflow/pkg/scheme"
	"github.com/siyul-park/uniflow/pkg/transaction"
)

// RDBNode represents a node for interacting with a relational database.
type RDBNode struct {
	*node.OneToOneNode
	db  *sqlx.DB
	txs *transaction.Local[*sqlx.Tx]
	mu  sync.RWMutex
}

// RDBNodeSpec holds the specifications for creating a RDBNode.
type RDBNodeSpec struct {
	scheme.SpecMeta `map:",inline"`
	Driver          string `map:"driver"`
	Source          string `map:"source"`
}

const KindRDB = "rdb"

// NewRDBNode creates a new RDBNode.
func NewRDBNode(db *sqlx.DB) *RDBNode {
	n := &RDBNode{
		db:  db,
		txs: transaction.NewLocal[*sqlx.Tx](),
	}
	n.OneToOneNode = node.NewOneToOneNode(n.action)

	return n
}

// Close closes resource associated with the node.
func (n *RDBNode) Close() error {
	n.mu.Lock()
	defer n.mu.Unlock()

	if err := n.OneToOneNode.Close(); err != nil {
		return err
	}
	return n.db.Close()
}

func (n *RDBNode) action(proc *process.Process, inPck *packet.Packet) (*packet.Packet, *packet.Packet) {
	n.mu.RLock()
	defer n.mu.RUnlock()

	ctx := proc.Context()

	query, ok := primitive.Pick[string](inPck.Payload())
	if !ok {
		query, ok = primitive.Pick[string](inPck.Payload(), "0")
	}
	if !ok {
		return nil, packet.WithError(packet.ErrInvalidPacket, inPck)
	}

	parent := proc.Transaction(inPck)
	tx, err := n.txs.LoadOrStore(parent, func() (*sqlx.Tx, error) {
		tx, err := n.db.BeginTxx(ctx, &sql.TxOptions{})
		if err != nil {
			return nil, err
		}

		parent.AddCommitHook(tx)
		parent.AddRollbackHook(tx)

		return tx, nil
	})
	if err != nil {
		return nil, packet.WithError(err, inPck)
	}

	stmt, err := tx.PrepareNamedContext(ctx, query)
	if err != nil {
		return nil, packet.WithError(err, inPck)
	}

	var rows *sqlx.Rows
	if len(stmt.Params) == 0 {
		args, _ := primitive.Pick[[]any](inPck.Payload(), "1")
		if rows, err = tx.QueryxContext(ctx, query, args...); err != nil {
			return nil, packet.WithError(err, inPck)
		}
	} else {
		var args any
		var ok bool
		args, ok = primitive.Pick[map[string]any](inPck.Payload(), "1")
		if !ok {
			args, _ = primitive.Pick[[]map[string]any](inPck.Payload(), "1")
		}
		if rows, err = stmt.QueryxContext(ctx, args); err != nil {
			return nil, packet.WithError(err, inPck)
		}
	}

	defer rows.Close()

	var results []map[string]any
	for rows.Next() {
		result := make(map[string]any)
		if err := rows.MapScan(result); err != nil {
			return nil, packet.WithError(err, inPck)
		}
		results = append(results, result)
	}

	outPayload, err := primitive.MarshalText(results)
	if err != nil {
		return nil, packet.WithError(err, inPck)
	}

	return packet.New(outPayload), nil
}

// NewRDBNodeCodec creates a new codec for RDBNodeSpec.
func NewRDBNodeCodec() scheme.Codec {
	return scheme.CodecWithType(func(spec *RDBNodeSpec) (node.Node, error) {
		db, err := sqlx.Connect(spec.Driver, spec.Source)
		if err != nil {
			return nil, err
		}
		return NewRDBNode(db), nil
	})
}
