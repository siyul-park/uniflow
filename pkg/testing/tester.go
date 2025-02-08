package testing

import (
	"time"

	"github.com/gofrs/uuid"
	"github.com/siyul-park/uniflow/pkg/process"
)

// Tester represents a test execution environment.
type Tester struct {
	*process.Process
	id        string
	name      string
	startTime time.Time
	endTime   time.Time
	assert    *Assertion
}

// NewTester creates a new Tester instance with the given name.
func NewTester(name string) *Tester {
	id, _ := uuid.NewV4()
	t := &Tester{
		Process:   process.New(),
		id:        id.String(),
		name:      name,
		startTime: time.Now(),
	}
	t.assert = NewAssertion(t)
	return t
}

// ID returns the unique identifier of the test.
func (t *Tester) ID() string {
	return t.id
}

// Name returns the name of the test.
func (t *Tester) Name() string {
	return t.name
}

// StartTime returns when the test started.
func (t *Tester) StartTime() time.Time {
	return t.startTime
}

// EndTime returns when the test ended.
func (t *Tester) EndTime() time.Time {
	return t.endTime
}

// Close marks the test as complete and records the end time.
func (t *Tester) Close(err error) {
	t.endTime = time.Now()
	t.Process.Exit(err)
}

// Assert returns the assertion helper for this test.
func (t *Tester) Assert() *Assertion {
	return t.assert
}
