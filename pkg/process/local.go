package process

import "sync"

// Local provides a concurrent cache for process-specific data.
type Local[T any] struct {
	data       map[*Process]T
	storeHooks map[*Process]StoreHooks[T]
	mu         sync.RWMutex
}

// NewLocal creates a new Local cache instance.
func NewLocal[T any]() *Local[T] {
	return &Local[T]{
		data:       make(map[*Process]T),
		storeHooks: make(map[*Process]StoreHooks[T]),
	}
}

// AddStoreHook adds a store hook for the given process.
func (l *Local[T]) AddStoreHook(proc *Process, hook StoreHook[T]) bool {
	l.mu.Lock()
	defer l.mu.Unlock()

	if _, ok := l.data[proc]; ok {
		l.mu.Unlock()
		hook.Store(l.data[proc])
		l.mu.Lock()
		return true
	}

	for _, h := range l.storeHooks[proc] {
		if h == hook {
			return false
		}
	}
	l.storeHooks[proc] = append(l.storeHooks[proc], hook)
	return true
}

// RemoveStoreHook removes a specific store hook for the given process.
func (l *Local[T]) RemoveStoreHook(proc *Process, hook StoreHook[T]) bool {
	l.mu.Lock()
	defer l.mu.Unlock()

	hooks, ok := l.storeHooks[proc]
	if !ok {
		return false
	}

	for i, h := range hooks {
		if h == hook {
			l.storeHooks[proc] = append(hooks[:i], hooks[i+1:]...)
			return true
		}
	}
	return false
}

// Keys returns all processes in the cache.
func (l *Local[T]) Keys() []*Process {
	l.mu.RLock()
	defer l.mu.RUnlock()

	keys := make([]*Process, 0, len(l.data))
	for proc := range l.data {
		keys = append(keys, proc)
	}
	return keys
}

// Load retrieves the value for the given process.
func (l *Local[T]) Load(proc *Process) (T, bool) {
	l.mu.RLock()
	defer l.mu.RUnlock()

	val, ok := l.data[proc]
	return val, ok
}

// Store sets the value for the given process.
func (l *Local[T]) Store(proc *Process, val T) {
	l.mu.Lock()

	_, ok := l.data[proc]

	l.data[proc] = val
	if !ok {
		proc.AddExitHook(ExitFunc(func(err error) {
			l.Delete(proc)
		}))
	}

	storeHooks := l.storeHooks[proc]
	delete(l.storeHooks, proc)

	l.mu.Unlock()

	storeHooks.Store(val)
}

// Delete removes the process and its data from the cache.
func (l *Local[T]) Delete(proc *Process) bool {
	l.mu.Lock()
	defer l.mu.Unlock()

	_, ok := l.data[proc]

	delete(l.data, proc)
	delete(l.storeHooks, proc)

	return ok
}

// LoadOrStore retrieves or stores a value for the given process.
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

	storeHooks := l.storeHooks[proc]
	delete(l.storeHooks, proc)

	l.mu.Unlock()

	storeHooks.Store(v)

	l.mu.Lock()

	return v, nil
}

// Close clears all cached data and hooks.
func (l *Local[T]) Close() {
	l.mu.Lock()
	defer l.mu.Unlock()

	l.data = make(map[*Process]T)
	l.storeHooks = make(map[*Process]StoreHooks[T])
}
