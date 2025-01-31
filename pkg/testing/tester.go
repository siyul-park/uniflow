package testing

import (
	"github.com/siyul-park/uniflow/pkg/process"
)

// Tester represents a tester with a name and an associated process.
type Tester struct {
	name string
	proc *process.Process
}

// NewTester creates a new Tester with the given name and a new process.
func NewTester(name string) *Tester {
	return &Tester{name: name, proc: process.New()}
}

// Name returns the name of the tester.
func (t *Tester) Name() string {
	return t.name
}

// Process returns the associated process.
func (t *Tester) Process() *process.Process {
	return t.proc
}

// Close exits the process with the given error.
func (t *Tester) Close(err error) {
	t.proc.Exit(err)
}
