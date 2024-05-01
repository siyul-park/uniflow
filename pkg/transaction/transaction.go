package transaction

import "sync"

type Transaction struct {
	commitHooks   []CommitHook
	rollbackHooks []RollbackHook
	mu            sync.RWMutex
}

var _ CommitHook = (*Transaction)(nil)
var _ RollbackHook = (*Transaction)(nil)

func New() *Transaction {
	return &Transaction{}
}

func (t *Transaction) AddCommitHook(hook CommitHook) {
	t.mu.Lock()
	defer t.mu.Unlock()

	t.commitHooks = append(t.commitHooks, hook)
}

func (t *Transaction) AddRollbackHook(hook RollbackHook) {
	t.mu.Lock()
	defer t.mu.Unlock()

	t.rollbackHooks = append(t.rollbackHooks, hook)
}

func (t *Transaction) Commit() error {
	t.mu.Lock()
	defer t.mu.Unlock()

	return t.commit()
}

func (t *Transaction) Rollback() error {
	t.mu.Lock()
	defer t.mu.Unlock()

	return t.rollback()
}

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
