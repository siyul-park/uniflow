package language

import (
	"context"
	"reflect"
	"time"
)

// Program represents an interface for running a compiled program with a given environment.
type Program interface {
	Run(context.Context, ...any) (any, error)
}

type program struct {
	fn func(context.Context, ...any) (any, error)
}

var _ Program = (*program)(nil)

// Timeout returns a Program that runs the given program with a specified timeout.
func Timeout(program Program, timeout time.Duration) Program {
	return RunFunc(func(ctx context.Context, args ...any) (any, error) {
		if timeout != 0 {
			var cancel func()
			ctx, cancel = context.WithTimeout(ctx, timeout)
			defer cancel()
		}
		return program.Run(ctx, args...)
	})
}

// Predicate returns a function that runs the program and checks if the first result is non-zero.
func Predicate[T any](program Program) func(context.Context, T) (bool, error) {
	return func(ctx context.Context, input T) (bool, error) {
		res, err := program.Run(ctx, input)
		if err != nil {
			return false, err
		}
		return !reflect.ValueOf(res).IsZero(), nil
	}
}

// Function returns a function that runs the program and returns the first result cast to type R.
func Function[T any, R any](program Program) func(context.Context, T) (R, error) {
	return func(ctx context.Context, input T) (R, error) {
		res, err := program.Run(ctx, input)
		r, _ := res.(R)
		return r, err
	}
}

// BiFunction returns a function that runs the program with two inputs and returns the first result cast to type R.
func BiFunction[T any, U any, R any](program Program) func(context.Context, T, U) (R, error) {
	return func(ctx context.Context, t T, u U) (R, error) {
		res, err := program.Run(ctx, t, u)
		r, _ := res.(R)
		return r, err
	}
}

// TriFunction returns a function that runs the program with three inputs and returns the first result cast to type R.
func TriFunction[T any, U any, V any, R any](program Program) func(context.Context, T, U, V) (R, error) {
	return func(ctx context.Context, t T, u U, v V) (R, error) {
		res, err := program.Run(ctx, t, u, v)
		r, _ := res.(R)
		return r, err
	}
}

// RunFunc creates a new Program implementation using the provided function.
func RunFunc(fn func(context.Context, ...any) (any, error)) Program {
	return &program{fn: fn}
}

func (p *program) Run(ctx context.Context, args ...any) (any, error) {
	return p.fn(ctx, args...)
}
