package language

// Program represents an interface for running a compiled program with a given environment.
type Program interface {
	Run(...any) (any, error)
}

// RunFunc is a function type that implements the Program interface.
type RunFunc func(...any) (any, error)

var _ Program = RunFunc(nil)

// Run executes the program with the provided environment using the RunFunc.
func (f RunFunc) Run(args ...any) (any, error) {
	return f(args...)
}
