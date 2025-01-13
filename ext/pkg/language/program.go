package language

import (
	"context"
	"reflect"
	"time"
)

// Program represents an interface for running a compiled program with a given environment.
type Program interface {
	Run(context.Context, []any) ([]any, error)
}

// RunFunc is a function type that implements the Program interface.
type RunFunc func(context.Context, []any) ([]any, error)

var _ Program = RunFunc(nil)

// Timeout returns a Program that runs the given program with a specified timeout.
func Timeout(program Program, timeout time.Duration) Program {
	return RunFunc(func(ctx context.Context, args []any) ([]any, error) {
		if timeout != 0 {
			var cancel func()
			ctx, cancel = context.WithTimeout(ctx, timeout)
			defer cancel()
		}
		return program.Run(ctx, args)
	})
}

// Predicate returns a function that runs the program and checks if the first result is non-zero.
func Predicate[T any](program Program) func(context.Context, T) (bool, error) {
	return func(ctx context.Context, input T) (bool, error) {
		res, err := program.Run(ctx, []any{input})
		if err != nil || len(res) == 0 {
			return false, err
		}
		return !reflect.ValueOf(res[0]).IsZero(), nil
	}
}

// Function returns a function that runs the program and returns the first result cast to type R.
func Function[T any, R any](program Program) func(context.Context, T) (R, error) {
	return func(ctx context.Context, input T) (R, error) {
		res, err := program.Run(ctx, []any{input})
		if err != nil || len(res) == 0 {
			var zero R
			return zero, err
		}
		return res[0].(R), nil
	}
}

// BiFunction returns a function that runs the program with two inputs and returns the first result cast to type R.
func BiFunction[T any, U any, R any](program Program) func(context.Context, T, U) (R, error) {
	return func(ctx context.Context, t T, u U) (R, error) {
		res, err := program.Run(ctx, []any{t, u})
		if err != nil || len(res) == 0 {
			var zero R
			return zero, err
		}
		return res[0].(R), nil
	}
}

// TriFunction returns a function that runs the program with three inputs and returns the first result cast to type R.
func TriFunction[T any, U any, V any, R any](program Program) func(context.Context, T, U, V) (R, error) {
	return func(ctx context.Context, t T, u U, v V) (R, error) {
		res, err := program.Run(ctx, []any{t, u, v})
		if err != nil || len(res) == 0 {
			var zero R
			return zero, err
		}
		return res[0].(R), nil
	}
}

// Run executes the program with the provided environment using the RunFunc.
func (f RunFunc) Run(ctx context.Context, args []any) ([]any, error) {
	return f(ctx, args)
}
