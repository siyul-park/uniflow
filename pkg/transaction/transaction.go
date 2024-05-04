package transaction

import "sync"

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
	t.mu.Lock()
	defer t.mu.Unlock()

	return t.commit()
}

// Rollback rolls back the transaction by executing all rollback hooks.
func (t *Transaction) Rollback() error {
	t.mu.Lock()
	defer t.mu.Unlock()

	return t.rollback()
}

// commit executes all commit hooks of the transaction.
func (t *Transaction) commit() error {
	defer func() {
		t.commitHooks = nil
		t.rollbackHooks = nil
	}()

	for _, hook := range t.commitHooks {
		if err := hook.Commit(); err != nil {
			_ = t.rollback()
			return err
		}
	}
	return nil
}

// rollback executes all rollback hooks of the transaction.
func (t *Transaction) rollback() error {
	defer func() {
		t.commitHooks = nil
		t.rollbackHooks = nil
	}()

	for i := len(t.rollbackHooks) - 1; i >= 0; i-- {
		hook := t.rollbackHooks[i]
		if err := hook.Rollback(); err != nil {
			return err
		}
	}
	return nil
}
