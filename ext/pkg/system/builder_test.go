package system

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/go-faker/faker/v4"
	"github.com/siyul-park/uniflow/pkg/hook"
	"github.com/siyul-park/uniflow/pkg/scheme"
	"github.com/siyul-park/uniflow/pkg/spec"
	"github.com/siyul-park/uniflow/pkg/symbol"
	"github.com/stretchr/testify/require"
)

func TestAddToHook(t *testing.T) {
	h := hook.New()

	err := AddToHook().AddToHook(h)
	require.NoError(t, err)

	n := NewSignalNode(nil)
	defer n.Close()

	sb := &symbol.Symbol{
		Spec: &spec.Meta{},
		Node: n,
	}

	err = h.Load(sb)
	require.NoError(t, err)

	err = h.Unload(sb)
	require.NoError(t, err)
}

func TestAddToScheme(t *testing.T) {
	s := scheme.New()

	err := AddToScheme().AddToScheme(s)
	require.NoError(t, err)

	tests := []string{KindSyscall, KindSignal}

	for _, tt := range tests {
		t.Run(tt, func(t *testing.T) {
			require.NotNil(t, s.KnownType(tt))
			require.NotNil(t, s.Codec(tt))
		})
	}
}

func TestSchemeRegister_Signal(t *testing.T) {
	t.Run("func() <-chan any", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.TODO(), time.Second)
		defer cancel()

		topic := faker.UUIDHyphenated()

		register := AddToScheme()

		err := register.SetSignal(topic, func() <-chan any {
			return make(chan any)
		})
		require.NoError(t, err)

		signal := register.Signal(topic)
		require.NotNil(t, signal)

		sig, err := signal(ctx)
		require.NoError(t, err)
		require.NotNil(t, sig)
	})

	t.Run("func() (<-chan any, error)", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.TODO(), time.Second)
		defer cancel()

		topic := faker.UUIDHyphenated()

		register := AddToScheme()

		err := register.SetSignal(topic, func() (<-chan any, error) {
			return make(chan any), nil
		})
		require.NoError(t, err)

		signal := register.Signal(topic)
		require.NotNil(t, signal)

		sig, err := signal(ctx)
		require.NoError(t, err)
		require.NotNil(t, sig)
	})

	t.Run("func(context.Context) <-chan any", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.TODO(), time.Second)
		defer cancel()

		topic := faker.UUIDHyphenated()

		register := AddToScheme()

		err := register.SetSignal(topic, func(_ context.Context) <-chan any {
			return make(chan any)
		})
		require.NoError(t, err)

		signal := register.Signal(topic)
		require.NotNil(t, signal)

		sig, err := signal(ctx)
		require.NoError(t, err)
		require.NotNil(t, sig)
	})

	t.Run("func(context.Context) (<-chan any, error)", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.TODO(), time.Second)
		defer cancel()

		topic := faker.UUIDHyphenated()

		register := AddToScheme()

		err := register.SetSignal(topic, func(_ context.Context) (<-chan any, error) {
			return make(chan any), nil
		})
		require.NoError(t, err)

		signal := register.Signal(topic)
		require.NotNil(t, signal)

		sig, err := signal(ctx)
		require.NoError(t, err)
		require.NotNil(t, sig)
	})
}

func TestSchemeRegister_Call(t *testing.T) {
	t.Run("func() void", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.TODO(), time.Second)
		defer cancel()

		opcode := faker.UUIDHyphenated()

		register := AddToScheme()

		err := register.SetCall(opcode, func() {})
		require.NoError(t, err)

		fn := register.Call(opcode)
		require.NotNil(t, fn)

		res, err := fn(ctx, nil)
		require.NoError(t, err)
		require.Len(t, res, 0)
	})

	t.Run("func() error", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.TODO(), time.Second)
		defer cancel()

		opcode := faker.UUIDHyphenated()

		register := AddToScheme()

		err := register.SetCall(opcode, func() error {
			return errors.New(faker.UUIDHyphenated())
		})
		require.NoError(t, err)

		fn := register.Call(opcode)
		require.NotNil(t, fn)

		_, err = fn(ctx, nil)
		require.Error(t, err)
	})

	t.Run("func(string) (string)", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.TODO(), time.Second)
		defer cancel()

		opcode := faker.UUIDHyphenated()

		register := AddToScheme()

		err := register.SetCall(opcode, func(arg string) string {
			return arg
		})
		require.NoError(t, err)

		fn := register.Call(opcode)
		require.NotNil(t, fn)

		arg := faker.UUIDHyphenated()

		res, err := fn(ctx, []any{arg})
		require.NoError(t, err)
		require.Len(t, res, 1)
		require.Equal(t, res[0], arg)
	})

	t.Run("func(string) (string, error)", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.TODO(), time.Second)
		defer cancel()

		opcode := faker.UUIDHyphenated()

		register := AddToScheme()

		err := register.SetCall(opcode, func(arg string) (string, error) {
			return "", errors.New(faker.UUIDHyphenated())
		})
		require.NoError(t, err)

		fn := register.Call(opcode)
		require.NotNil(t, fn)

		arg := faker.UUIDHyphenated()

		_, err = fn(ctx, []any{arg})
		require.Error(t, err)
	})

	t.Run("func(context.Context, string) (string)", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.TODO(), time.Second)
		defer cancel()

		opcode := faker.UUIDHyphenated()

		register := AddToScheme()

		err := register.SetCall(opcode, func(_ context.Context, arg string) string {
			return arg
		})
		require.NoError(t, err)

		fn := register.Call(opcode)
		require.NotNil(t, fn)

		arg := faker.UUIDHyphenated()

		res, err := fn(ctx, []any{arg})
		require.NoError(t, err)
		require.Len(t, res, 1)
		require.Equal(t, res[0], arg)
	})

	t.Run("func(context.Context, string) (string, error)", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.TODO(), time.Second)
		defer cancel()

		opcode := faker.UUIDHyphenated()

		register := AddToScheme()

		err := register.SetCall(opcode, func(_ context.Context, arg string) (string, error) {
			return "", errors.New(faker.UUIDHyphenated())
		})
		require.NoError(t, err)

		fn := register.Call(opcode)
		require.NotNil(t, fn)

		arg := faker.UUIDHyphenated()

		_, err = fn(ctx, []any{arg})
		require.Error(t, err)
	})

	t.Run("func(string, string) (string, string)", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.TODO(), time.Second)
		defer cancel()

		opcode := faker.UUIDHyphenated()

		register := AddToScheme()

		err := register.SetCall(opcode, func(arg1, arg2 string) (string, string) {
			return arg1, arg2
		})
		require.NoError(t, err)

		fn := register.Call(opcode)
		require.NotNil(t, fn)

		arg := faker.UUIDHyphenated()

		res, err := fn(ctx, []any{arg, arg})
		require.NoError(t, err)
		require.Len(t, res, 2)
		require.Equal(t, res[0], arg)
		require.Equal(t, res[1], arg)
	})

	t.Run("func(string, string) (string, string, error)", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.TODO(), time.Second)
		defer cancel()

		opcode := faker.UUIDHyphenated()

		register := AddToScheme()

		err := register.SetCall(opcode, func(arg1, arg2 string) (string, string, error) {
			return "", "", errors.New(faker.UUIDHyphenated())
		})
		require.NoError(t, err)

		fn := register.Call(opcode)
		require.NotNil(t, fn)

		arg := faker.UUIDHyphenated()

		_, err = fn(ctx, []any{arg, arg})
		require.Error(t, err)
	})
}
