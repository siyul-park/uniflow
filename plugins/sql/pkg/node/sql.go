package node

import (
	"database/sql"
	"sync"

	"github.com/jmoiron/sqlx"
	"github.com/siyul-park/uniflow/pkg/node"
	"github.com/siyul-park/uniflow/pkg/packet"
	"github.com/siyul-park/uniflow/pkg/process"
	"github.com/siyul-park/uniflow/pkg/scheme"
	"github.com/siyul-park/uniflow/pkg/spec"
	"github.com/siyul-park/uniflow/pkg/types"
)

// SQLNodeSpec defines the specifications for creating a SQLNode.
type SQLNodeSpec struct {
	spec.Meta `json:",inline"`
	Driver    string             `json:"driver" validate:"required"`
	Source    string             `json:"source,omitempty"`
	Isolation sql.IsolationLevel `json:"isolation,omitempty"`
}

// SQLNode represents a node for interacting with a relational database.
type SQLNode struct {
	*node.OneToOneNode
	db        *sqlx.DB
	txs       *process.Local[*sqlx.Tx]
	isolation sql.IsolationLevel
	mu        sync.RWMutex
}

const KindSQL = "sql"

// NewSQLNodeCodec creates a new codec for SQLNodeSpec.
func NewSQLNodeCodec() scheme.Codec {
	return scheme.CodecWithType(func(spec *SQLNodeSpec) (node.Node, error) {
		db, err := sqlx.Connect(spec.Driver, spec.Source)
		if err != nil {
			return nil, err
		}

		n := NewSQLNode(db)
		n.SetIsolation(spec.Isolation)
		return n, nil
	})
}

// NewSQLNode creates a new SQLNode.
func NewSQLNode(db *sqlx.DB) *SQLNode {
	n := &SQLNode{
		db:  db,
		txs: process.NewLocal[*sqlx.Tx](),
	}
	n.OneToOneNode = node.NewOneToOneNode(n.action)
	return n
}

// Isolation returns the isolation level of the SQLNode.
func (n *SQLNode) Isolation() sql.IsolationLevel {
	n.mu.RLock()
	defer n.mu.RUnlock()

	return n.isolation
}

// SetIsolation sets the isolation level of the SQLNode.
func (n *SQLNode) SetIsolation(isolation sql.IsolationLevel) {
	n.mu.Lock()
	defer n.mu.Unlock()

	n.isolation = isolation
}

// Close closes meta associated with the node.
func (n *SQLNode) Close() error {
	n.mu.RLock()
	defer n.mu.RUnlock()

	if err := n.OneToOneNode.Close(); err != nil {
		return err
	}
	return n.db.Close()
}

func (n *SQLNode) action(proc *process.Process, inPck *packet.Packet) (*packet.Packet, *packet.Packet) {
	n.mu.RLock()
	defer n.mu.RUnlock()

	query, err := types.Cast[string](inPck.Payload())
	if err != nil {
		if query, err = types.Cast[string](types.Lookup(inPck.Payload(), 0)); err != nil {
			return nil, packet.New(types.NewError(err))
		}
	}

	tx, err := n.txs.LoadOrStore(proc, func() (*sqlx.Tx, error) {
		tx, err := n.db.BeginTxx(proc, &sql.TxOptions{
			Isolation: n.isolation,
		})
		if err != nil {
			return nil, err
		}

		proc.AddExitHook(process.ExitFunc(func(err error) {
			if err != nil {
				_ = tx.Rollback()
			} else {
				_ = tx.Commit()
			}
		}))

		return tx, nil
	})
	if err != nil {
		return nil, packet.New(types.NewError(err))
	}

	stmt, err := tx.PrepareNamedContext(proc, query)
	if err != nil {
		return nil, packet.New(types.NewError(err))
	}
	defer stmt.Close()

	var rows *sqlx.Rows
	if len(stmt.Params) == 0 {
		args, _ := types.Cast[[]any](types.Lookup(inPck.Payload(), 1), nil)
		if rows, err = tx.QueryxContext(proc, query, args...); err != nil {
			return nil, packet.New(types.NewError(err))
		}
	} else {
		var arg any
		var err error
		arg, err = types.Cast[map[string]any](types.Lookup(inPck.Payload(), 1), nil)
		if err != nil {
			arg, _ = types.Cast[[]map[string]any](types.Lookup(inPck.Payload(), 1), nil)
		}

		query, args, err := tx.BindNamed(query, arg)
		if err != nil {
			return nil, packet.New(types.NewError(err))
		}

		if rows, err = tx.QueryxContext(proc, query, args...); err != nil {
			return nil, packet.New(types.NewError(err))
		}
	}
	defer rows.Close()

	var results []map[string]any
	for rows.Next() {
		result := make(map[string]any)
		if err := rows.MapScan(result); err != nil {
			return nil, packet.New(types.NewError(err))
		}
		for k, v := range result {
			if v == nil {
				delete(result, k)
			}
		}
		results = append(results, result)
	}

	outPayload, err := types.Marshal(results)
	if err != nil {
		return nil, packet.New(types.NewError(err))
	}
	return packet.New(outPayload), nil
}
