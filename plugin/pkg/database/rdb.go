package database

import (
	"context"
	"database/sql"
	"sync"

	"github.com/jmoiron/sqlx"
	"github.com/siyul-park/uniflow/pkg/node"
	"github.com/siyul-park/uniflow/pkg/packet"
	"github.com/siyul-park/uniflow/pkg/primitive"
	"github.com/siyul-park/uniflow/pkg/process"
)

type RDBNode struct {
	*node.OneToOneNode
	db  *sqlx.DB
	txs *process.Local
	mu  sync.RWMutex
}

func NewRDBNode(db *sqlx.DB) *RDBNode {
	n := &RDBNode{
		db:  db,
		txs: process.NewLocal(),
	}
	n.OneToOneNode = node.NewOneToOneNode(n.action)

	return n
}

func (n *RDBNode) action(proc *process.Process, inPck *packet.Packet) (*packet.Packet, *packet.Packet) {
	n.mu.RLock()
	defer n.mu.RUnlock()

	ctx, cancel := context.WithCancel(context.Background())
	go func() {
		<-proc.Done()
		cancel()
	}()

	query, ok := primitive.Pick[string](inPck.Payload())
	if !ok {
		query, ok = primitive.Pick[string](inPck.Payload(), "0")
	}
	args, _ := primitive.Pick[[]any](inPck.Payload(), "1")

	if !ok {
		return nil, packet.WithError(packet.ErrInvalidPacket, inPck)
	}

	val, err := n.txs.LoadOrStore(proc, func() (any, error) {
		tx, err := n.db.BeginTxx(ctx, &sql.TxOptions{})
		if err != nil {
			return nil, err
		}

		go func() {
			<-proc.Done()

			if proc.Err() == nil {
				_ = tx.Commit()
			} else {
				_ = tx.Rollback()
			}
		}()

		return tx, nil
	})
	if err != nil {
		return nil, packet.WithError(err, inPck)
	}

	tx := val.(*sqlx.Tx)

	rows, err := tx.QueryxContext(ctx, query, args...)
	if err != nil {
		return nil, packet.WithError(err, inPck)
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

func (n *RDBNode) Close() error {
	n.mu.Lock()
	defer n.mu.Unlock()

	if err := n.OneToOneNode.Close(); err != nil {
		return err
	}
	n.txs.Close()
	return nil
}
