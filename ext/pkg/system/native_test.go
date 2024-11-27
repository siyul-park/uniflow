package system

import (
	"context"
	"github.com/pkg/errors"
	"testing"
	"time"

	"github.com/go-faker/faker/v4"
	"github.com/siyul-park/uniflow/pkg/node"
	"github.com/siyul-park/uniflow/pkg/packet"
	"github.com/siyul-park/uniflow/pkg/port"
	"github.com/siyul-park/uniflow/pkg/process"
	"github.com/siyul-park/uniflow/pkg/types"
	"github.com/stretchr/testify/assert"
)

func TestNativeNodeCodec_Compile(t *testing.T) {
	opcode := faker.UUIDHyphenated()

	codec := NewNativeNodeCodec(map[string]any{
		opcode: func() {},
	})

	spec := &NativeNodeSpec{
		OPCode: opcode,
	}

	n, err := codec.Compile(spec)
	assert.NoError(t, err)
	assert.NotNil(t, n)
	assert.NoError(t, n.Close())
}

func TestNativeNodeCodec_Load(t *testing.T) {
	t.Run("func() void", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.TODO(), time.Second)
		defer cancel()

		opcode := faker.UUIDHyphenated()

		codec := NewNativeNodeCodec(map[string]any{
			opcode: func() {},
		})

		fn, err := codec.Load(opcode)
		assert.NoError(t, err)

		res, err := fn(ctx, nil)
		assert.NoError(t, err)
		assert.Len(t, res, 0)
	})

	t.Run("func() error", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.TODO(), time.Second)
		defer cancel()

		opcode := faker.UUIDHyphenated()

		codec := NewNativeNodeCodec(map[string]any{
			opcode: func() error {
				return errors.New(faker.Word())
			},
		})

		fn, err := codec.Load(opcode)
		assert.NoError(t, err)

		_, err = fn(ctx, nil)
		assert.Error(t, err)
	})

	t.Run("func(string) (string)", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.TODO(), time.Second)
		defer cancel()

		opcode := faker.UUIDHyphenated()

		codec := NewNativeNodeCodec(map[string]any{
			opcode: func(arg string) string {
				return arg
			},
		})

		fn, err := codec.Load(opcode)
		assert.NoError(t, err)

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

		codec := NewNativeNodeCodec(map[string]any{
			opcode: func(arg string) (string, error) {
				return "", errors.New(faker.Word())
			},
		})

		fn, err := codec.Load(opcode)
		assert.NoError(t, err)

		arg := faker.UUIDHyphenated()

		_, err = fn(ctx, []any{arg})
		assert.Error(t, err)
	})

	t.Run("func(context.Context, string) (string)", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.TODO(), time.Second)
		defer cancel()

		opcode := faker.UUIDHyphenated()

		codec := NewNativeNodeCodec(map[string]any{
			opcode: func(_ context.Context, arg string) string {
				return arg
			},
		})

		fn, err := codec.Load(opcode)
		assert.NoError(t, err)

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

		codec := NewNativeNodeCodec(map[string]any{
			opcode: func(_ context.Context, arg string) (string, error) {
				return "", errors.New(faker.Word())
			},
		})

		fn, err := codec.Load(opcode)
		assert.NoError(t, err)

		arg := faker.UUIDHyphenated()

		_, err = fn(ctx, []any{arg})
		assert.Error(t, err)
	})

	t.Run("func(string, string) (string, string)", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.TODO(), time.Second)
		defer cancel()

		opcode := faker.UUIDHyphenated()

		codec := NewNativeNodeCodec(map[string]any{
			opcode: func(arg1, arg2 string) (string, string) {
				return arg1, arg2
			},
		})

		fn, err := codec.Load(opcode)
		assert.NoError(t, err)

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

		codec := NewNativeNodeCodec(map[string]any{
			opcode: func(arg1, arg2 string) (string, string, error) {
				return "", "", errors.New(faker.Word())
			},
		})

		fn, err := codec.Load(opcode)
		assert.NoError(t, err)

		arg := faker.UUIDHyphenated()

		_, err = fn(ctx, []any{arg, arg})
		assert.Error(t, err)
	})
}

func TestNewNativeNode(t *testing.T) {
	n, err := NewNativeNode(nil)
	assert.NoError(t, err)
	assert.NotNil(t, n)
	assert.NoError(t, n.Close())
}

func TestNativeNode_SendAndReceive(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.TODO(), time.Second)
	defer cancel()

	n, _ := NewNativeNode(func(ctx context.Context, arguments []any) ([]any, error) {
		return arguments, nil
	})
	defer n.Close()

	in := port.NewOut()
	in.Link(n.In(node.PortIn))

	proc := process.New()
	defer proc.Exit(nil)

	inWriter := in.Open(proc)

	inPayload := types.NewString(faker.UUIDHyphenated())
	inPck := packet.New(inPayload)

	inWriter.Write(inPck)

	select {
	case outPck := <-inWriter.Receive():
		assert.Equal(t, inPayload, outPck.Payload())
	case <-ctx.Done():
		assert.Fail(t, ctx.Err().Error())
	}
}

func BenchmarkNativeNode_SendAndReceive(b *testing.B) {
	n, _ := NewNativeNode(func(ctx context.Context, arguments []any) ([]any, error) {
		return arguments, nil
	})
	defer n.Close()

	in := port.NewOut()
	in.Link(n.In(node.PortIn))

	proc := process.New()
	defer proc.Exit(nil)

	inWriter := in.Open(proc)

	inPayload := types.NewString(faker.UUIDHyphenated())
	inPck := packet.New(inPayload)

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		inWriter.Write(inPck)
		<-inWriter.Receive()
	}
}
