package transaction

import "sync"

type Local[T any] struct {
	data map[*Transaction]T
	done chan struct{}
	mu   sync.RWMutex
}

func NewLocal[T any]() *Local[T] {
	return &Local[T]{
		data: make(map[*Transaction]T),
		done: make(chan struct{}),
	}
}

func (l *Local[T]) Load(tx *Transaction) (T, bool) {
	l.mu.RLock()
	defer l.mu.RUnlock()

	val, ok := l.data[tx]
	return val, ok
}

func (l *Local[T]) Store(tx *Transaction, val T) {
	l.mu.Lock()
	defer l.mu.Unlock()

	_, ok := l.data[tx]

	l.data[tx] = val

	if !ok {
		tx.AddCommitHook(CommitHookFunc(func() error {
			l.Delete(tx)
			return nil
		}))
		tx.AddRollbackHook(RollbackHookFunc(func() error {
			l.Delete(tx)
			return nil
		}))
	}
}

func (l *Local[T]) Delete(tx *Transaction) bool {
	l.mu.Lock()
	defer l.mu.Unlock()

	if _, ok := l.data[tx]; !ok {
		return false
	}
	delete(l.data, tx)
	return true
}

// LoadOrStore retrieves the value associated with a process, or stores a new value if the process is not present.
func (l *Local[T]) LoadOrStore(tx *Transaction, val func() (T, error)) (T, error) {
	l.mu.Lock()
	defer l.mu.Unlock()

	if v, ok := l.data[tx]; ok {
		return v, nil
	}

	v, err := val()
	if err != nil {
		return v, err
	}

	l.data[tx] = v

	tx.AddCommitHook(CommitHookFunc(func() error {
		l.Delete(tx)
		return nil
	}))
	tx.AddRollbackHook(RollbackHookFunc(func() error {
		l.Delete(tx)
		return nil
	}))

	return v, nil
}

func (l *Local[T]) Close() {
	l.mu.Lock()
	defer l.mu.Unlock()

	select {
	case <-l.done:
		return
	default:
	}

	l.data = make(map[*Transaction]T)
	close(l.done)
}
