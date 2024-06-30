package process

import "sync"

// Local represents a local cache for storing data associated with processes.
type Local[T any] struct {
	data map[*Process]T
	mu   sync.RWMutex
}

// NewLocal creates and initializes a new Local cache.
func NewLocal[T any]() *Local[T] {
	return &Local[T]{
		data: make(map[*Process]T),
	}
}

// Load retrieves the value associated with a given process.
func (l *Local[T]) Load(proc *Process) (T, bool) {
	l.mu.RLock()
	defer l.mu.RUnlock()

	val, ok := l.data[proc]
	return val, ok
}

// Store stores a value associated with a process in the cache.
func (l *Local[T]) Store(proc *Process, val T) {
	l.mu.Lock()
	defer l.mu.Unlock()

	_, ok := l.data[proc]
	l.data[proc] = val
	if !ok {
		proc.AtExit(ExitHookFunc(func(err error) {
			l.Delete(proc)
		}))
	}
}

// Delete removes the association of a process and its value from the cache.
func (l *Local[T]) Delete(proc *Process) bool {
	l.mu.Lock()
	defer l.mu.Unlock()

	if _, ok := l.data[proc]; !ok {
		return false
	}
	delete(l.data, proc)
	return true
}

// LoadOrStore retrieves the value associated with a process, or stores a new value if the process is not present.
func (l *Local[T]) LoadOrStore(proc *Process, val func() (T, error)) (T, error) {
	l.mu.Lock()
	defer l.mu.Unlock()

	if v, ok := l.data[proc]; ok {
		return v, nil
	}

	v, err := val()
	if err != nil {
		return v, err
	}

	l.data[proc] = v
	proc.AtExit(ExitHookFunc(func(err error) {
		l.Delete(proc)
	}))

	return v, nil
}

// Close closes the Local cache, releasing associated resources.
func (l *Local[T]) Close() {
	l.mu.Lock()
	defer l.mu.Unlock()

	l.data = make(map[*Process]T)
}
