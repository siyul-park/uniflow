package process

import "sync"

// Local provides a concurrent cache for storing data associated with processes.
type Local[T any] struct {
	data    map[*Process]T
	watches map[*Process][]func(T)
	mu      sync.RWMutex
}

// NewLocal creates and returns a new Local cache instance.
func NewLocal[T any]() *Local[T] {
	return &Local[T]{
		data:    make(map[*Process]T),
		watches: make(map[*Process][]func(T)),
	}
}

// Watch adds a watcher for the specified process.
func (l *Local[T]) Watch(proc *Process, watch func(T)) {
	l.mu.Lock()

	v, exists := l.data[proc]
	if !exists {
		l.watches[proc] = append(l.watches[proc], watch)
	}

	l.mu.Unlock()

	if exists {
		watch(v)
	}
}

// Keys returns a slice of all processes in the cache.
func (l *Local[T]) Keys() []*Process {
	l.mu.RLock()
	defer l.mu.RUnlock()

	keys := make([]*Process, 0, len(l.data))
	for proc := range l.data {
		keys = append(keys, proc)
	}
	return keys
}

// Load retrieves the value associated with the specified process.
func (l *Local[T]) Load(proc *Process) (T, bool) {
	l.mu.RLock()
	defer l.mu.RUnlock()

	val, exists := l.data[proc]
	return val, exists
}

// Store associates a value with the specified process.
func (l *Local[T]) Store(proc *Process, val T) {
	l.mu.Lock()

	_, exists := l.data[proc]

	l.data[proc] = val
	if !exists {
		proc.AddExitHook(ExitFunc(func(err error) {
			l.Delete(proc)
		}))
	}

	watches := l.watches[proc]
	delete(l.watches, proc)

	l.mu.Unlock()

	for _, watch := range watches {
		watch(val)
	}
}

// Delete removes the specified process and its associated value from the cache.
func (l *Local[T]) Delete(proc *Process) bool {
	l.mu.Lock()
	defer l.mu.Unlock()

	_, exists := l.data[proc]

	delete(l.data, proc)
	delete(l.watches, proc)

	return exists
}

// LoadOrStore retrieves the value for the specified process or stores a new value if absent.
func (l *Local[T]) LoadOrStore(proc *Process, val func() (T, error)) (T, error) {
	l.mu.Lock()
	defer l.mu.Unlock()

	if v, exists := l.data[proc]; exists {
		return v, nil
	}

	v, err := val()
	if err != nil {
		return v, err
	}

	l.data[proc] = v
	proc.AddExitHook(ExitFunc(func(err error) {
		l.Delete(proc)
	}))

	watches := l.watches[proc]
	delete(l.watches, proc)

	l.mu.Unlock()

	for _, watch := range watches {
		watch(v)
	}

	l.mu.Lock()

	return v, nil
}

// Close clears the entire cache, removing all processes and their values.
func (l *Local[T]) Close() {
	l.mu.Lock()
	defer l.mu.Unlock()

	l.data = make(map[*Process]T)
	l.watches = make(map[*Process][]func(T))
}
