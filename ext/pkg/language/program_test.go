package language

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestTimeout(t *testing.T) {
	program := RunFunc(func(ctx context.Context, args ...any) (any, error) {
		_, ok := ctx.Deadline()
		assert.True(t, ok)
		return nil, nil
	})
	timeout := Timeout(program, 1*time.Second)

	_, err := timeout.Run(context.Background(), nil)
	assert.NoError(t, err)
}

func TestPredicate(t *testing.T) {
	program := RunFunc(func(ctx context.Context, args ...any) (any, error) {
		return []any{1}, nil
	})
	predicate := Predicate[int](program)

	result, err := predicate(context.Background(), 1)
	assert.NoError(t, err)
	assert.True(t, result)
}

func TestFunction(t *testing.T) {
	program := RunFunc(func(ctx context.Context, args ...any) (any, error) {
		return "result", nil
	})
	function := Function[int, string](program)

	result, err := function(context.Background(), 1)
	assert.NoError(t, err)
	assert.Equal(t, "result", result)
}

func TestBiFunction(t *testing.T) {
	program := RunFunc(func(ctx context.Context, args ...any) (any, error) {
		return "result", nil
	})
	biFunction := BiFunction[int, int, string](program)

	result, err := biFunction(context.Background(), 1, 2)
	assert.NoError(t, err)
	assert.Equal(t, "result", result)
}

func TestTriFunction(t *testing.T) {
	program := RunFunc(func(ctx context.Context, args ...any) (any, error) {
		return "result", nil
	})
	triFunction := TriFunction[int, int, int, string](program)

	result, err := triFunction(context.Background(), 1, 2, 3)
	assert.NoError(t, err)
	assert.Equal(t, "result", result)
}
