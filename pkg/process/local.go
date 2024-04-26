package process

import "sync"

// Local represents a local cache for storing data associated with processes.
type Local struct {
	data map[*Process]any
	done chan struct{}
	mu   sync.RWMutex
}

// NewLocal creates and initializes a new Local cache.
func NewLocal() *Local {
	return &Local{
		data: make(map[*Process]any),
		done: make(chan struct{}),
	}
}

// Load retrieves the value associated with a given process.
func (l *Local) Load(proc *Process) (any, bool) {
	l.mu.RLock()
	defer l.mu.RUnlock()

	val, ok := l.data[proc]
	return val, ok
}

// Store stores a value associated with a process in the cache.
func (l *Local) Store(proc *Process, val any) {
	l.mu.Lock()
	defer l.mu.Unlock()

	_, ok := l.data[proc]

	l.data[proc] = val

	if !ok {
		go func() {
			select {
			case <-proc.Done():
				l.Delete(proc)
			case <-l.done:
				return
			}
		}()
	}
}

// Delete removes the association of a process and its value from the cache.
func (l *Local) Delete(proc *Process) bool {
	l.mu.Lock()
	defer l.mu.Unlock()

	if _, ok := l.data[proc]; !ok {
		return false
	}
	delete(l.data, proc)
	return true
}

// LoadOrStore retrieves the value associated with a process, or stores a new value if the process is not present.
func (l *Local) LoadOrStore(proc *Process, val func() (any, error)) (any, error) {
	l.mu.Lock()
	defer l.mu.Unlock()

	if v, ok := l.data[proc]; ok {
		return v, nil
	}

	v, err := val()
	if err != nil {
		return nil, err
	}

	l.data[proc] = v

	go func() {
		select {
		case <-proc.Done():
			l.Delete(proc)
		case <-l.done:
			return
		}
	}()

	return v, nil
}

// Close closes the Local cache, releasing associated resources.
func (l *Local) Close() {
	l.mu.Lock()
	defer l.mu.Unlock()

	select {
	case <-l.done:
		return
	default:
	}

	l.data = make(map[*Process]any)
	close(l.done)
}
