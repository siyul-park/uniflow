package process

import (
	"math"
	"sync"

	"github.com/siyul-park/uniflow/pkg/packet"
	"github.com/siyul-park/uniflow/pkg/transaction"
)

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

func (t *Transactions) Commit() error {
	tx := t.Get(nil)
	return tx.Commit()
}

func (t *Transactions) Rollback() error {
	tx := t.Get(nil)
	return tx.Rollback()
}

func (t *Transactions) Get(pck *packet.Packet) *transaction.Transaction {
	t.mu.RLock()
	defer t.mu.RUnlock()

	return t.lookup(pck)
}

func (t *Transactions) Set(pck *packet.Packet, tx *transaction.Transaction) {
	t.mu.Lock()
	defer t.mu.Unlock()

	parent := t.lookup(pck)
	parent.AddCommitHook(transaction.CommitHookFunc(func() error {
		t.Delete(pck)
		return tx.Commit()
	}))
	parent.AddRollbackHook(transaction.RollbackHookFunc(func() error {
		t.Delete(pck)
		return tx.Rollback()
	}))

	t.transactions[pck] = tx
}

func (t *Transactions) Delete(pck *packet.Packet) {
	t.mu.Lock()
	defer t.mu.Unlock()

	delete(t.transactions, pck)
}

func (t *Transactions) lookup(pck *packet.Packet) *transaction.Transaction {
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
