package process

import (
	"sync"
	"sync/atomic"
)

// Local provides a concurrent cache for process-specific eager.
type Local[T any] struct {
	eager      map[*Process]T
	lazy       map[*Process]*lazy[T]
	storeHooks map[*Process]StoreHooks[T]
	mu         sync.RWMutex
}

type lazy[T any] struct {
	fn    func() (T, error)
	value T
	error error
	done  atomic.Uint32
	mu    sync.Mutex
}

// NewLocal creates a new Local cache instance.
func NewLocal[T any]() *Local[T] {
	return &Local[T]{
		eager:      make(map[*Process]T),
		lazy:       make(map[*Process]*lazy[T]),
		storeHooks: make(map[*Process]StoreHooks[T]),
	}
}

// AddStoreHook adds a store hook for the given process.
func (l *Local[T]) AddStoreHook(proc *Process, hook StoreHook[T]) bool {
	l.mu.Lock()
	defer l.mu.Unlock()

	if val, ok := l.eager[proc]; ok {
		l.mu.Unlock()
		defer l.mu.Lock()

		hook.Store(val)
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

	keys := make([]*Process, 0, len(l.eager))
	for proc := range l.eager {
		keys = append(keys, proc)
	}
	return keys
}

// Load retrieves the value for the given process.
func (l *Local[T]) Load(proc *Process) (T, bool) {
	l.mu.RLock()
	defer l.mu.RUnlock()

	val, ok := l.eager[proc]
	return val, ok
}

// Store sets the value for the given process.
func (l *Local[T]) Store(proc *Process, val T) {
	l.mu.Lock()

	_, ok := l.eager[proc]

	l.eager[proc] = val
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

// Delete removes the process and its eager from the cache.
func (l *Local[T]) Delete(proc *Process) bool {
	l.mu.Lock()
	defer l.mu.Unlock()

	_, ok := l.eager[proc]

	delete(l.eager, proc)
	delete(l.storeHooks, proc)

	return ok
}

// LoadOrStore retrieves or stores a value for the given process.
func (l *Local[T]) LoadOrStore(proc *Process, val func() (T, error)) (T, error) {
	l.mu.RLock()
	v, ok := l.eager[proc]
	l.mu.RUnlock()
	if ok {
		return v, nil
	}

	l.mu.Lock()

	if v, ok := l.eager[proc]; ok {
		l.mu.Unlock()
		return v, nil
	}

	fn, ok := l.lazy[proc]
	if !ok {
		fn = &lazy[T]{fn: val}
		l.lazy[proc] = fn
	}

	l.mu.Unlock()

	v, err := fn.Do()
	if err != nil {
		return v, err
	}

	l.mu.Lock()

	l.eager[proc] = v
	delete(l.lazy, proc)

	storeHooks := l.storeHooks[proc]
	delete(l.storeHooks, proc)

	l.mu.Unlock()

	proc.AddExitHook(ExitFunc(func(err error) {
		l.Delete(proc)
	}))

	storeHooks.Store(v)

	return v, nil
}

// Close clears all cached eager and hooks.
func (l *Local[T]) Close() {
	l.mu.Lock()
	defer l.mu.Unlock()

	l.eager = make(map[*Process]T)
	l.lazy = make(map[*Process]*lazy[T])
	l.storeHooks = make(map[*Process]StoreHooks[T])
}

func (o *lazy[T]) Do() (T, error) {
	o.mu.Lock()
	defer o.mu.Unlock()

	if o.done.Load() == 0 {
		defer o.done.Store(1)
		o.value, o.error = o.fn()
	}
	return o.value, o.error
}
