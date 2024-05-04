package process

import (
	"math"
	"sync"

	"github.com/siyul-park/uniflow/pkg/packet"
	"github.com/siyul-park/uniflow/pkg/transaction"
)

// Transactions manages transactions associated with packet packets.
type Transactions struct {
	stack        *Stack
	transactions map[*packet.Packet]*transaction.Transaction
	mu           sync.RWMutex
}

var _ transaction.CommitHook = (*Transactions)(nil)
var _ transaction.RollbackHook = (*Transactions)(nil)

func newTransactions(stack *Stack) *Transactions {
	return &Transactions{
		stack: stack,
		transactions: map[*packet.Packet]*transaction.Transaction{
			nil: transaction.New(),
		},
	}
}

// Commit commits the transaction.
func (t *Transactions) Commit() error {
	tx := t.Get(nil)
	return tx.Commit()
}

// Rollback rolls back the transaction.
func (t *Transactions) Rollback() error {
	tx := t.Get(nil)
	return tx.Rollback()
}

// Get returns the transaction associated with the packet.
func (t *Transactions) Get(pck *packet.Packet) *transaction.Transaction {
	t.mu.RLock()
	defer t.mu.RUnlock()

	return t.lookup(pck)
}

// Set associates the transaction with the packet.
func (t *Transactions) Set(pck *packet.Packet, tx *transaction.Transaction) {
	t.mu.Lock()
	defer t.mu.Unlock()

	parent := t.lookup(pck)
	t.transactions[pck] = tx

	tx.AddCommitHook(transaction.CommitHookFunc(func() error {
		t.delete(pck, tx)
		return nil
	}))
	tx.AddRollbackHook(transaction.RollbackHookFunc(func() error {
		t.delete(pck, tx)
		return nil
	}))

	parent.AddCommitHook(tx)
	parent.AddRollbackHook(tx)

	if t.stack.Has(nil, pck) {
		go func() {
			<-t.stack.Done(pck)
			_ = tx.Commit()
		}()
	}
}

func (t *Transactions) lookup(pck *packet.Packet) *transaction.Transaction {
	if tx, ok := t.transactions[pck]; ok {
		return tx
	}

	cost := math.MaxInt
	tx := t.transactions[nil]

	for k, v := range t.transactions {
		if c := t.stack.Cost(k, pck); c <= cost {
			tx = v
			cost = c
		}
	}

	return tx
}

func (t *Transactions) delete(pck *packet.Packet, tx *transaction.Transaction) {
	t.mu.Lock()
	defer t.mu.Unlock()

	if cur := t.transactions[pck]; cur == tx {
		delete(t.transactions, pck)
	}
}
