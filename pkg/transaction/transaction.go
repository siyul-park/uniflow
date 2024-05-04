package transaction

import (
	"sync"
)

// Transaction represents a transaction with commit and rollback functionality.
type Transaction struct {
	commitHooks   []CommitHook
	rollbackHooks []RollbackHook
	mu            sync.RWMutex
}

var _ CommitHook = (*Transaction)(nil)
var _ RollbackHook = (*Transaction)(nil)

// New creates a new Transaction instance.
func New() *Transaction {
	return &Transaction{}
}

// AddCommitHook adds a commit hook to the transaction.
func (t *Transaction) AddCommitHook(hook CommitHook) {
	t.mu.Lock()
	defer t.mu.Unlock()

	t.commitHooks = append(t.commitHooks, hook)
}

// AddRollbackHook adds a rollback hook to the transaction.
func (t *Transaction) AddRollbackHook(hook RollbackHook) {
	t.mu.Lock()
	defer t.mu.Unlock()

	t.rollbackHooks = append(t.rollbackHooks, hook)
}

// Commit commits the transaction by executing all commit hooks.
func (t *Transaction) Commit() error {
	commitHooks, rollbackHooks := t.hooks()

	for _, hook := range commitHooks {
		if err := hook.Commit(); err != nil {
			for i := len(rollbackHooks) - 1; i >= 0; i-- {
				hook := rollbackHooks[i]
				_ = hook.Rollback()
			}
			return err
		}
	}

	return nil
}

// Rollback rolls back the transaction by executing all rollback hooks.
func (t *Transaction) Rollback() error {
	_, rollbackHooks := t.hooks()

	var err error
	for i := len(rollbackHooks) - 1; i >= 0; i-- {
		hook := rollbackHooks[i]
		if e := hook.Rollback(); e != nil {
			err = e
		}
	}

	return err
}

func (t *Transaction) hooks() ([]CommitHook, []RollbackHook) {
	t.mu.Lock()
	defer t.mu.Unlock()

	defer func() {
		t.commitHooks = nil
		t.rollbackHooks = nil
	}()

	return t.commitHooks, t.rollbackHooks
}
