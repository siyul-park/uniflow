package process

import "sync"

// Local is a concurrent cache for storing data associated with processes.
type Local[T any] struct {
	data    map[*Process]T
	watches map[*Process][]func(T)
	mu      sync.RWMutex
}

// NewLocal creates a new Local cache instance.
func NewLocal[T any]() *Local[T] {
	return &Local[T]{
		data:    make(map[*Process]T),
		watches: make(map[*Process][]func(T)),
	}
}

// Watch adds a watcher function for a process.
// If the process already has a value, the watcher is called immediately.
func (l *Local[T]) Watch(proc *Process, watch func(T)) bool {
	l.mu.Lock()

	val, ok := l.data[proc]
	if !ok {
		l.watches[proc] = append(l.watches[proc], watch)
	}

	l.mu.Unlock()

	if ok {
		watch(val)
	}
	return ok
}

// Keys returns all processes in the Local cache.
func (l *Local[T]) Keys() []*Process {
	l.mu.RLock()
	defer l.mu.RUnlock()

	keys := make([]*Process, 0, len(l.data))
	for proc := range l.data {
		keys = append(keys, proc)
	}
	return keys
}

// Load retrieves the value for a given process.
func (l *Local[T]) Load(proc *Process) (T, bool) {
	l.mu.RLock()
	defer l.mu.RUnlock()

	val, ok := l.data[proc]
	return val, ok
}

// Store stores a value for a process and sets an exit hook if new.
func (l *Local[T]) Store(proc *Process, val T) {
	l.mu.Lock()

	_, ok := l.data[proc]
	l.data[proc] = val

	if !ok {
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

// Delete removes a process and its value from the Local cache.
func (l *Local[T]) Delete(proc *Process) bool {
	l.mu.Lock()
	defer l.mu.Unlock()

	_, ok := l.data[proc]

	delete(l.data, proc)
	delete(l.watches, proc)

	return ok
}

// LoadOrStore retrieves the value for a process or stores a new value if not present.
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
	proc.AddExitHook(ExitFunc(func(err error) {
		l.Delete(proc)
	}))

	return v, nil
}

// Close clears the Local cache.
func (l *Local[T]) Close() {
	l.mu.Lock()
	defer l.mu.Unlock()

	l.data = make(map[*Process]T)
	l.watches = make(map[*Process][]func(T))
}
