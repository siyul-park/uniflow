package process

import (
	"testing"

	"github.com/siyul-park/uniflow/pkg/packet"
	"github.com/siyul-park/uniflow/pkg/transaction"
	"github.com/stretchr/testify/assert"
)

func TestTransactions_Get(t *testing.T) {
	s := newStack()
	defer s.Close()

	tx := newTransactions(s)

	pck1 := packet.New(nil)
	pck2 := packet.New(nil)

	assert.Equal(t, tx.Get(nil), tx.Get(pck1))
	assert.Equal(t, tx.Get(nil), tx.Get(pck2))

	s.Add(pck1, pck2)
	assert.Equal(t, tx.Get(nil), tx.Get(pck1))
	assert.Equal(t, tx.Get(nil), tx.Get(pck2))

	tx1 := transaction.New()
	tx2 := transaction.New()

	tx.Set(pck1, tx1)
	assert.Equal(t, tx1, tx.Get(pck1))
	assert.Equal(t, tx1, tx.Get(pck2))

	tx.Set(pck2, tx2)
	assert.Equal(t, tx1, tx.Get(pck1))
	assert.Equal(t, tx2, tx.Get(pck2))
}

func TestTransactions_Commit(t *testing.T) {
	s := newStack()
	defer s.Close()

	tx := newTransactions(s)

	pck1 := packet.New(nil)
	tx1 := transaction.New()

	s.Add(nil, pck1)
	tx.Set(pck1, tx1)

	count := 0
	tx1.AddCommitHook(transaction.CommitHookFunc(func() error {
		count += 1
		return nil
	}))

	err := tx.Commit()
	assert.NoError(t, err)
	assert.Equal(t, 1, count)
}

func TestTransactions_Rollback(t *testing.T) {
	s := newStack()
	defer s.Close()

	tx := newTransactions(s)

	pck1 := packet.New(nil)
	tx1 := transaction.New()

	s.Add(nil, pck1)
	tx.Set(pck1, tx1)

	count := 0
	tx1.AddRollbackHook(transaction.RollbackHookFunc(func() error {
		count += 1
		return nil
	}))

	err := tx.Rollback()
	assert.NoError(t, err)
	assert.Equal(t, 1, count)
}
