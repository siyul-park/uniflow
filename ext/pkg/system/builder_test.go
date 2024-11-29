package system

import (
	"context"
	"errors"
	"github.com/go-faker/faker/v4"
	"github.com/siyul-park/uniflow/pkg/hook"
	"github.com/siyul-park/uniflow/pkg/scheme"
	"github.com/siyul-park/uniflow/pkg/spec"
	"github.com/siyul-park/uniflow/pkg/symbol"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func TestAddToHook(t *testing.T) {
	h := hook.New()

	err := AddToHook().AddToHook(h)
	assert.NoError(t, err)

	n := NewSignalNode(nil)
	defer n.Close()

	sb := &symbol.Symbol{
		Spec: &spec.Meta{},
		Node: n,
	}

	err = h.Load(sb)
	assert.NoError(t, err)

	err = h.Unload(sb)
	assert.NoError(t, err)
}

func TestAddToScheme(t *testing.T) {
	s := scheme.New()

	err := AddToScheme(nil).AddToScheme(s)
	assert.NoError(t, err)

	tests := []string{KindNative, KindSignal}

	for _, tt := range tests {
		t.Run(tt, func(t *testing.T) {
			assert.NotNil(t, s.KnownType(tt))
			assert.NotNil(t, s.Codec(tt))
		})
	}
}

func TestSchemeRegister_Signal(t *testing.T) {
	t.Run("func() <-chan any", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.TODO(), time.Second)
		defer cancel()

		opcode := faker.UUIDHyphenated()

		register := AddToScheme(map[string]any{
			opcode: func() <-chan any {
				return make(chan any)
			},
		})

		signal := register.Signal(opcode)
		assert.NotNil(t, signal)

		sig, err := signal(ctx)
		assert.NoError(t, err)
		assert.NotNil(t, sig)
	})

	t.Run("func() (<-chan any, error)", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.TODO(), time.Second)
		defer cancel()

		opcode := faker.UUIDHyphenated()

		register := AddToScheme(map[string]any{
			opcode: func() (<-chan any, error) {
				return make(chan any), nil
			},
		})

		signal := register.Signal(opcode)
		assert.NotNil(t, signal)

		sig, err := signal(ctx)
		assert.NoError(t, err)
		assert.NotNil(t, sig)
	})

	t.Run("func(context.Context) <-chan any", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.TODO(), time.Second)
		defer cancel()

		opcode := faker.UUIDHyphenated()

		register := AddToScheme(map[string]any{
			opcode: func(_ context.Context) <-chan any {
				return make(chan any)
			},
		})

		signal := register.Signal(opcode)
		assert.NotNil(t, signal)

		sig, err := signal(ctx)
		assert.NoError(t, err)
		assert.NotNil(t, sig)
	})

	t.Run("func(context.Context) (<-chan any, error)", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.TODO(), time.Second)
		defer cancel()

		opcode := faker.UUIDHyphenated()

		register := AddToScheme(map[string]any{
			opcode: func(_ context.Context) (<-chan any, error) {
				return make(chan any), nil
			},
		})

		signal := register.Signal(opcode)
		assert.NotNil(t, signal)

		sig, err := signal(ctx)
		assert.NoError(t, err)
		assert.NotNil(t, sig)
	})
}

func TestSchemeRegister_Function(t *testing.T) {
	t.Run("func() void", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.TODO(), time.Second)
		defer cancel()

		opcode := faker.UUIDHyphenated()

		register := AddToScheme(map[string]any{
			opcode: func() {},
		})

		fn := register.Function(opcode)
		assert.NotNil(t, fn)

		res, err := fn(ctx, nil)
		assert.NoError(t, err)
		assert.Len(t, res, 0)
	})

	t.Run("func() error", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.TODO(), time.Second)
		defer cancel()

		opcode := faker.UUIDHyphenated()

		register := AddToScheme(map[string]any{
			opcode: func() error {
				return errors.New(faker.Word())
			},
		})

		fn := register.Function(opcode)
		assert.NotNil(t, fn)

		_, err := fn(ctx, nil)
		assert.Error(t, err)
	})

	t.Run("func(string) (string)", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.TODO(), time.Second)
		defer cancel()

		opcode := faker.UUIDHyphenated()

		register := AddToScheme(map[string]any{
			opcode: func(arg string) string {
				return arg
			},
		})

		fn := register.Function(opcode)
		assert.NotNil(t, fn)

		arg := faker.UUIDHyphenated()

		res, err := fn(ctx, []any{arg})
		assert.NoError(t, err)
		assert.Len(t, res, 1)
		assert.Equal(t, res[0], arg)
	})

	t.Run("func(string) (string, error)", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.TODO(), time.Second)
		defer cancel()

		opcode := faker.UUIDHyphenated()

		register := AddToScheme(map[string]any{
			opcode: func(arg string) (string, error) {
				return "", errors.New(faker.Word())
			},
		})

		fn := register.Function(opcode)
		assert.NotNil(t, fn)

		arg := faker.UUIDHyphenated()

		_, err := fn(ctx, []any{arg})
		assert.Error(t, err)
	})

	t.Run("func(context.Context, string) (string)", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.TODO(), time.Second)
		defer cancel()

		opcode := faker.UUIDHyphenated()

		register := AddToScheme(map[string]any{
			opcode: func(_ context.Context, arg string) string {
				return arg
			},
		})

		fn := register.Function(opcode)
		assert.NotNil(t, fn)

		arg := faker.UUIDHyphenated()

		res, err := fn(ctx, []any{arg})
		assert.NoError(t, err)
		assert.Len(t, res, 1)
		assert.Equal(t, res[0], arg)
	})

	t.Run("func(context.Context, string) (string, error)", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.TODO(), time.Second)
		defer cancel()

		opcode := faker.UUIDHyphenated()

		register := AddToScheme(map[string]any{
			opcode: func(_ context.Context, arg string) (string, error) {
				return "", errors.New(faker.Word())
			},
		})

		fn := register.Function(opcode)
		assert.NotNil(t, fn)

		arg := faker.UUIDHyphenated()

		_, err := fn(ctx, []any{arg})
		assert.Error(t, err)
	})

	t.Run("func(string, string) (string, string)", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.TODO(), time.Second)
		defer cancel()

		opcode := faker.UUIDHyphenated()

		register := AddToScheme(map[string]any{
			opcode: func(arg1, arg2 string) (string, string) {
				return arg1, arg2
			},
		})

		fn := register.Function(opcode)
		assert.NotNil(t, fn)

		arg := faker.UUIDHyphenated()

		res, err := fn(ctx, []any{arg, arg})
		assert.NoError(t, err)
		assert.Len(t, res, 2)
		assert.Equal(t, res[0], arg)
		assert.Equal(t, res[1], arg)
	})

	t.Run("func(string, string) (string, string, error)", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.TODO(), time.Second)
		defer cancel()

		opcode := faker.UUIDHyphenated()

		register := AddToScheme(map[string]any{
			opcode: func(arg1, arg2 string) (string, string, error) {
				return "", "", errors.New(faker.Word())
			},
		})

		fn := register.Function(opcode)
		assert.NotNil(t, fn)

		arg := faker.UUIDHyphenated()

		_, err := fn(ctx, []any{arg, arg})
		assert.Error(t, err)
	})
}
