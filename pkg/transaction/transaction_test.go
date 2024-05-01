package transaction

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestTransaction_Commit(t *testing.T) {
	tx := New()

	count := 0
	tx.AddCommitHook(CommitHookFunc(func() error {
		count += 1
		return nil
	}))

	err := tx.Commit()
	assert.NoError(t, err)
	assert.Equal(t, 1, count)
}

func TestTransaction_Rollback(t *testing.T) {
	tx := New()

	count := 0
	tx.AddRollbackHook(RollbackHookFunc(func() error {
		count += 1
		return nil
	}))

	err := tx.Rollback()
	assert.NoError(t, err)
	assert.Equal(t, 1, count)
}
