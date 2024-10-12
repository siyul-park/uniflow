package language

import "context"

// Program represents an interface for running a compiled program with a given environment.
type Program interface {
	Run(context.Context, ...any) (any, error)
}

// RunFunc is a function type that implements the Program interface.
type RunFunc func(context.Context, ...any) (any, error)

var _ Program = RunFunc(nil)

// Run executes the program with the provided environment using the RunFunc.
func (f RunFunc) Run(ctx context.Context, args ...any) (any, error) {
	return f(ctx, args...)
}
