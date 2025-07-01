package process

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/gofrs/uuid"
)

// Process represents a unit of execution with eager, status, and lifecycle management.
type Process struct {
	id        uuid.UUID
	data      map[any]any
	status    Status
	err       error
	startTime time.Time
	endTime   time.Time
	exitHooks ExitHooks
	done      chan struct{}
	wait      sync.WaitGroup
	mu        sync.RWMutex
	parent    *Process
}

// Status represents the current state of a process.
type Status int

const (
	StatusRunning Status = iota
	StatusTerminated
)

var (
	_ context.Context = (*Process)(nil)
	_ json.Marshaler  = (*Process)(nil)
)

// New creates and returns a new Process instance with an initial state.
func New() *Process {
	return &Process{
		id:        uuid.Must(uuid.NewV7()),
		data:      make(map[any]any),
		startTime: time.Now(),
		done:      make(chan struct{}),
	}
}

// ID returns the unique identifier of the process.
func (p *Process) ID() uuid.UUID {
	return p.id
}

// Keys returns a list of all keys associated with the process and its parent.
func (p *Process) Keys() []any {
	p.mu.RLock()
	defer p.mu.RUnlock()

	keys := make([]any, 0, len(p.data))
	for key := range p.data {
		keys = append(keys, key)
	}
	if p.parent != nil {
		keys = append(keys, p.parent.Keys()...)
	}
	return keys
}

// Value returns the value associated with the given key in the process or its parent.
func (p *Process) Value(key any) any {
	p.mu.RLock()
	defer p.mu.RUnlock()

	if val, ok := p.data[key]; ok {
		return val
	}
	if p.parent != nil {
		return p.parent.Value(key)
	}
	return nil
}

// SetValue stores a value in the process associated with the given key.
func (p *Process) SetValue(key, val any) {
	p.mu.Lock()
	defer p.mu.Unlock()

	p.data[key] = val
}

// RemoveValue removes the value associated with the given key from the process.
func (p *Process) RemoveValue(key any) any {
	p.mu.Lock()
	defer p.mu.Unlock()

	if val, ok := p.data[key]; ok {
		delete(p.data, key)
		return val
	}
	if p.parent != nil {
		return p.parent.RemoveValue(key)
	}
	return nil
}

// Status returns the current status of the process.
func (p *Process) Status() Status {
	p.mu.RLock()
	defer p.mu.RUnlock()

	return p.status
}

// Err returns the error associated with the process, if any.
func (p *Process) Err() error {
	p.mu.RLock()
	defer p.mu.RUnlock()

	if p.err != nil {
		return p.err
	}

	if p.status == StatusTerminated {
		return context.Canceled
	}
	return nil
}

// StartTime returns the time when the process started.
func (p *Process) StartTime() time.Time {
	return p.startTime
}

// EndTime returns the time when the process ended.
func (p *Process) EndTime() time.Time {
	p.mu.RLock()
	defer p.mu.RUnlock()

	return p.endTime
}

// Parent returns the parent process, if any.
func (p *Process) Parent() *Process {
	return p.parent
}

// Deadline always returns (time.Time{}, false) as there is no cancellation deadline in the process.
func (p *Process) Deadline() (time.Time, bool) {
	return time.Time{}, false
}

// Done returns a channel that is closed when the process is done.
func (p *Process) Done() <-chan struct{} {
	return p.done
}

// Join waits for all child processes to complete.
func (p *Process) Join() {
	p.wait.Wait()
}

// Fork creates a new child process that inherits data and context from the parent.
func (p *Process) Fork() *Process {
	p.wait.Add(1)

	child := &Process{
		id:      uuid.Must(uuid.NewV7()),
		data:    make(map[any]any),
		endTime: time.Now(),
		exitHooks: []ExitHook{
			ExitFunc(func(err error) {
				p.wait.Done()
			}),
		},
		done:   make(chan struct{}),
		parent: p,
	}
	p.AddExitHook(child)
	return child
}

// Exit terminates the process with the given error, executing all exit hooks.
func (p *Process) Exit(err error) {
	p.mu.Lock()
	exitHooks := p.exitHooks
	if p.status != StatusTerminated {
		close(p.done)

		p.data = make(map[any]any)
		p.status = StatusTerminated
		p.err = err
		p.endTime = time.Now()
		p.exitHooks = nil
	}
	p.mu.Unlock()

	exitHooks.Exit(err)
}

// AddExitHook adds a hook to be executed when the process terminates.
func (p *Process) AddExitHook(hook ExitHook) bool {
	p.mu.Lock()

	if p.status == StatusTerminated {
		err := p.err
		p.mu.Unlock()
		hook.Exit(err)
		return false
	}

	for _, h := range p.exitHooks {
		if h == hook {
			p.mu.Unlock()
			return false
		}
	}
	p.exitHooks = append(p.exitHooks, hook)

	p.mu.Unlock()

	return true
}

// MarshalJSON encodes the Process into a compact, standard-form JSON object.
func (p *Process) MarshalJSON() ([]byte, error) {
	p.mu.RLock()
	defer p.mu.RUnlock()

	data := map[string]any{
		"id":         p.id.String(),
		"status":     p.status,
		"start_time": p.startTime.Unix(),
	}

	if p.parent != nil {
		data["parent_id"] = p.parent.ID().String()
	}
	if !p.endTime.IsZero() {
		data["end_time"] = p.endTime.Unix()
	}
	if p.err != nil {
		data["error"] = p.err.Error()
	}

	for _, key := range p.Keys() {
		if val, ok := p.data[key]; ok {
			k := fmt.Sprint(key)
			if bs, err := json.Marshal(val); err != nil {
				data[k] = "<native>"
			} else {
				data[k] = json.RawMessage(bs)
			}
		}
	}

	return json.Marshal(data)
}
