package testing

import (
	"time"

	"github.com/gofrs/uuid"
	"github.com/siyul-park/uniflow/pkg/process"
)

// Tester manages a test process with an associated name.
type Tester struct {
	name string
	proc *process.Process
}

// NewTester creates a Tester with the given name and a new process.
func NewTester(name string) *Tester {
	return &Tester{name: name, proc: process.New()}
}

// ID returns the unique identifier of the test process.
func (t *Tester) ID() uuid.UUID {
	return t.proc.ID()
}

// Name returns the name of the tester.
func (t *Tester) Name() string {
	return t.name
}

// StartTime returns the start time of the test process.
func (t *Tester) StartTime() time.Time {
	return t.proc.StartTime()
}

// EndTime returns the end time of the test process.
func (t *Tester) EndTime() time.Time {
	return t.proc.EndTime()
}

// AddExitHook registers a function to execute when the process terminates.
func (t *Tester) AddExitHook(hook process.ExitHook) bool {
	return t.proc.AddExitHook(hook)
}

// Process returns the underlying test process.
func (t *Tester) Process() *process.Process {
	return t.proc
}

// Done returns a channel that closes when the test process completes.
func (t *Tester) Done() <-chan struct{} {
	return t.proc.Done()
}

// Close terminates the test process with the given error.
func (t *Tester) Close(err error) {
	t.proc.Exit(err)
}
