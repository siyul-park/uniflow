package language

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestTimeout(t *testing.T) {
	program := RunFunc(func(ctx context.Context, args ...any) (any, error) {
		_, ok := ctx.Deadline()
		require.True(t, ok)
		return nil, nil
	})
	timeout := Timeout(program, time.Second)

	_, err := timeout.Run(context.Background(), nil)
	require.NoError(t, err)
}

func TestPredicate(t *testing.T) {
	program := RunFunc(func(ctx context.Context, args ...any) (any, error) {
		return []any{1}, nil
	})
	predicate := Predicate[int](program)

	result, err := predicate(context.Background(), 1)
	require.NoError(t, err)
	require.True(t, result)
}

func TestFunction(t *testing.T) {
	program := RunFunc(func(ctx context.Context, args ...any) (any, error) {
		return "result", nil
	})
	function := Function[int, string](program)

	result, err := function(context.Background(), 1)
	require.NoError(t, err)
	require.Equal(t, "result", result)
}

func TestBiFunction(t *testing.T) {
	program := RunFunc(func(ctx context.Context, args ...any) (any, error) {
		return "result", nil
	})
	biFunction := BiFunction[int, int, string](program)

	result, err := biFunction(context.Background(), 1, 2)
	require.NoError(t, err)
	require.Equal(t, "result", result)
}

func TestTriFunction(t *testing.T) {
	program := RunFunc(func(ctx context.Context, args ...any) (any, error) {
		return "result", nil
	})
	triFunction := TriFunction[int, int, int, string](program)

	result, err := triFunction(context.Background(), 1, 2, 3)
	require.NoError(t, err)
	require.Equal(t, "result", result)
}
