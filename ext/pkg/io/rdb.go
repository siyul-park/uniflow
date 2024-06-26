package io

import (
	"database/sql"
	"sync"

	"github.com/jmoiron/sqlx"
	"github.com/siyul-park/uniflow/pkg/encoding"
	"github.com/siyul-park/uniflow/pkg/node"
	"github.com/siyul-park/uniflow/pkg/object"
	"github.com/siyul-park/uniflow/pkg/packet"
	"github.com/siyul-park/uniflow/pkg/process"
	"github.com/siyul-park/uniflow/pkg/scheme"
	"github.com/siyul-park/uniflow/pkg/spec"
)

// RDBNode represents a node for interacting with a relational database.
type RDBNode struct {
	*node.OneToOneNode
	db        *sqlx.DB
	txs       *process.Local[*sqlx.Tx]
	isolation sql.IsolationLevel
	mu        sync.RWMutex
}

// RDBNodeSpec holds the specifications for creating a RDBNode.
type RDBNodeSpec struct {
	spec.Meta `map:",inline"`
	Driver    string             `map:"driver"`
	Source    string             `map:"source"`
	Isolation sql.IsolationLevel `map:"isolation"`
}

const KindRDB = "rdb"

// NewRDBNode creates a new RDBNode.
func NewRDBNode(db *sqlx.DB) *RDBNode {
	n := &RDBNode{
		db:  db,
		txs: process.NewLocal[*sqlx.Tx](),
	}
	n.OneToOneNode = node.NewOneToOneNode(n.action)

	return n
}

// Isolation returns the isolation level of the RDBNode.
func (n *RDBNode) Isolation() sql.IsolationLevel {
	n.mu.RLock()
	defer n.mu.RUnlock()

	return n.isolation
}

// SetIsolation sets the isolation level of the RDBNode.
func (n *RDBNode) SetIsolation(isolation sql.IsolationLevel) {
	n.mu.Lock()
	defer n.mu.Unlock()

	n.isolation = isolation
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

	query, ok := object.Pick[string](inPck.Payload())
	if !ok {
		query, ok = object.Pick[string](inPck.Payload(), "0")
	}
	if !ok {
		return nil, packet.New(object.NewError(encoding.ErrUnsupportedValue))
	}

	tx, err := n.txs.LoadOrStore(proc, func() (*sqlx.Tx, error) {
		tx, err := n.db.BeginTxx(ctx, &sql.TxOptions{
			Isolation: n.isolation,
		})
		if err != nil {
			return nil, err
		}

		proc.AddExitHook(process.ExitHookFunc(func(err error) {
			if err != nil {
				tx.Rollback()
			} else {
				tx.Commit()
			}
		}))

		return tx, nil
	})
	if err != nil {
		return nil, packet.New(object.NewError(err))
	}

	stmt, err := tx.PrepareNamedContext(ctx, query)
	if err != nil {
		return nil, packet.New(object.NewError(err))
	}

	var rows *sqlx.Rows
	if len(stmt.Params) == 0 {
		args, _ := object.Pick[[]any](inPck.Payload(), "1")
		if rows, err = tx.QueryxContext(ctx, query, args...); err != nil {
			return nil, packet.New(object.NewError(err))
		}
	} else {
		var args any
		var ok bool
		args, ok = object.Pick[map[string]any](inPck.Payload(), "1")
		if !ok {
			args, _ = object.Pick[[]map[string]any](inPck.Payload(), "1")
		}
		if rows, err = stmt.QueryxContext(ctx, args); err != nil {
			return nil, packet.New(object.NewError(err))
		}
	}

	defer rows.Close()

	var results []map[string]any
	for rows.Next() {
		result := make(map[string]any)
		if err := rows.MapScan(result); err != nil {
			return nil, packet.New(object.NewError(err))
		}
		results = append(results, result)
	}

	outPayload, err := object.MarshalText(results)
	if err != nil {
		return nil, packet.New(object.NewError(err))
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

		n := NewRDBNode(db)
		n.SetIsolation(spec.Isolation)

		return n, nil
	})
}
