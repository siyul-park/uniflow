package testing

// Suite is an interface that runs a collection of tests.
type Suite interface {
	// Run executes the tests in the suite.
	Run(t *Tester)
}

type suite struct {
	run func(t *Tester)
}

var _ Suite = (*suite)(nil)

// RunFunc creates a new Suite with the provided run function.
func RunFunc(run func(t *Tester)) Suite {
	return &suite{run: run}
}

// Run executes the tests in the suite.
func (s *suite) Run(t *Tester) {
	s.run(t)
}
