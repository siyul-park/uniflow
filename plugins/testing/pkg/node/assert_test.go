package node

import (
	"context"
	"testing"
	"time"

	"github.com/go-faker/faker/v4"
	"github.com/gofrs/uuid"
	"github.com/pkg/errors"
	"github.com/siyul-park/uniflow/pkg/language/json"
	"github.com/siyul-park/uniflow/pkg/meta"
	"github.com/siyul-park/uniflow/pkg/node"
	"github.com/siyul-park/uniflow/pkg/packet"
	"github.com/siyul-park/uniflow/pkg/port"
	"github.com/siyul-park/uniflow/pkg/process"
	"github.com/siyul-park/uniflow/pkg/runtime"
	"github.com/siyul-park/uniflow/pkg/spec"
	"github.com/siyul-park/uniflow/pkg/types"
	"github.com/stretchr/testify/require"
)

func TestNewAssertNodeCodec_Compile(t *testing.T) {
	compiler := json.NewCompiler()
	agent := runtime.NewAgent()
	defer agent.Close()

	codec := NewAssertNodeCodec(compiler, agent)
	require.NotNil(t, codec)

	t.Run("Compile", func(t *testing.T) {
		spec := &AssertNodeSpec{
			Meta: spec.Meta{
				ID:        uuid.Must(uuid.NewV7()),
				Kind:      faker.UUIDHyphenated(),
				Namespace: meta.DefaultNamespace,
				Name:      faker.UUIDHyphenated(),
			},
			Expect: "{}",
		}

		n, err := codec.Compile(spec)
		require.NoError(t, err)
		require.NotNil(t, n)
		require.NoError(t, n.Close())
	})

	t.Run("WithTarget", func(t *testing.T) {
		spec := &AssertNodeSpec{
			Meta: spec.Meta{
				ID:        uuid.Must(uuid.NewV7()),
				Kind:      faker.UUIDHyphenated(),
				Namespace: meta.DefaultNamespace,
				Name:      faker.UUIDHyphenated(),
			},
			Expect: "{}",
			Target: &spec.Port{
				ID:   uuid.Must(uuid.NewV7()),
				Name: "target",
				Port: "out",
			},
			Timeout: time.Second,
		}

		n, err := codec.Compile(spec)
		require.NoError(t, err)
		require.NotNil(t, n)
		require.NoError(t, n.Close())
	})

	t.Run("CompileError", func(t *testing.T) {
		spec := &AssertNodeSpec{
			Meta: spec.Meta{
				ID:        uuid.Must(uuid.NewV7()),
				Kind:      faker.UUIDHyphenated(),
				Namespace: meta.DefaultNamespace,
				Name:      faker.UUIDHyphenated(),
			},
			Expect: "{ error }",
		}

		n, err := codec.Compile(spec)
		require.Error(t, err)
		require.Nil(t, n)
	})
}

func TestAssertNode_SetTarget(t *testing.T) {
	n := NewAssertNode(func(ctx context.Context, payload any) (bool, error) {
		return true, nil
	})
	defer n.Close()

	target := func(proc *process.Process, payload any, index int) (any, int, error) {
		return payload, index, nil
	}
	n.SetTarget(target)

	require.NotNil(t, n)
}

func TestAssertNode_Port(t *testing.T) {
	n := NewAssertNode(nil)
	defer n.Close()

	require.NotNil(t, n.In(node.PortIn))
	require.NotNil(t, n.Out(node.PortOut))
}

func TestAssertNode_SendAndReceive(t *testing.T) {
	t.Run("DirectAssert", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()

		assert := NewAssertNode(func(ctx context.Context, payload any) (bool, error) {
			if val, ok := payload.(int); ok {
				return val == 10, nil
			}
			return false, nil
		})
		defer assert.Close()

		in := port.NewOut()
		in.Link(assert.In(node.PortIn))

		out := port.NewIn()
		assert.Out(node.PortOut).Link(out)

		proc := process.New()
		defer proc.Exit(nil)

		inWriter := in.Open(proc)
		outReader := out.Open(proc)

		inPayload, err := types.Marshal([]any{10, -1})
		require.NoError(t, err)
		inPck := packet.New(inPayload)

		inWriter.Write(inPck)

		select {
		case outPck := <-outReader.Read():
			require.NotNil(t, outPck)
			outReader.Receive(outPck)
			require.Equal(t, inPayload, outPck.Payload())
		case <-ctx.Done():
			require.Fail(t, ctx.Err().Error())
		}
	})

	t.Run("AssertFail", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()

		assert := NewAssertNode(func(ctx context.Context, payload any) (bool, error) {
			if val, ok := payload.(int); ok {
				return val == 10, nil
			}
			return false, nil
		})
		defer assert.Close()

		in := port.NewOut()
		in.Link(assert.In(node.PortIn))

		out := port.NewIn()
		assert.Out(node.PortError).Link(out)

		proc := process.New()
		defer proc.Exit(nil)

		inWriter := in.Open(proc)
		outReader := out.Open(proc)

		inPayload, err := types.Marshal([]any{99, -1})
		require.NoError(t, err)
		inPck := packet.New(inPayload)

		inWriter.Write(inPck)

		select {
		case outPck := <-outReader.Read():
			require.NotNil(t, outPck)
			outReader.Receive(outPck)
			require.ErrorIs(t, outPck.Payload().(types.Error).Unwrap(), ErrAssertFail)
		case <-ctx.Done():
			require.Fail(t, ctx.Err().Error())
		}
	})

	t.Run("TargetAssert", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()

		assert := NewAssertNode(func(ctx context.Context, payload any) (bool, error) {
			if val, ok := payload.(int); ok {
				return val == 10, nil
			}
			return false, nil
		})
		defer assert.Close()

		in := port.NewOut()
		in.Link(assert.In(node.PortIn))

		out := port.NewIn()
		assert.Out(node.PortOut).Link(out)

		proc := process.New()
		defer proc.Exit(nil)

		inWriter := in.Open(proc)
		outReader := out.Open(proc)

		inPayload, err := types.Marshal([]any{10, -1})
		require.NoError(t, err)
		inPck := packet.New(inPayload)

		assert.SetTarget(func(proc *process.Process, payload any, index int) (any, int, error) {
			return payload, index, nil
		})

		inWriter.Write(inPck)

		select {
		case outPck := <-outReader.Read():
			require.NotNil(t, outPck)
			outReader.Receive(outPck)
			require.Equal(t, inPayload, outPck.Payload())
		case <-ctx.Done():
			require.Fail(t, ctx.Err().Error())
		}
	})

	t.Run("TargetNotFound", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()

		assert := NewAssertNode(func(ctx context.Context, payload any) (bool, error) {
			if val, ok := payload.(int); ok {
				return val == 10, nil
			}
			return false, nil
		})
		defer assert.Close()

		assert.SetTarget(func(proc *process.Process, payload any, index int) (any, int, error) {
			return nil, 0, errors.WithStack(ErrAssertFail)
		})

		in := port.NewOut()
		in.Link(assert.In(node.PortIn))

		out := port.NewIn()
		assert.Out(node.PortError).Link(out)

		proc := process.New()
		defer proc.Exit(nil)

		inWriter := in.Open(proc)
		outReader := out.Open(proc)

		inPayload, err := types.Marshal([]any{10, -1})
		require.NoError(t, err)
		inPck := packet.New(inPayload)

		inWriter.Write(inPck)

		select {
		case outPck := <-outReader.Read():
			require.NotNil(t, outPck)
			outReader.Receive(outPck)
			require.ErrorIs(t, outPck.Payload().(types.Error).Unwrap(), ErrAssertFail)
		case <-ctx.Done():
			require.Fail(t, ctx.Err().Error())
		}
	})

	t.Run("ExpectError", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()

		assert := NewAssertNode(func(ctx context.Context, payload any) (bool, error) {
			return false, errors.New(faker.Sentence())
		})
		defer assert.Close()

		in := port.NewOut()
		in.Link(assert.In(node.PortIn))

		out := port.NewIn()
		assert.Out(node.PortError).Link(out)

		proc := process.New()
		defer proc.Exit(nil)

		inWriter := in.Open(proc)
		outReader := out.Open(proc)

		inPayload, err := types.Marshal([]any{10, -1})
		require.NoError(t, err)
		inPck := packet.New(inPayload)

		inWriter.Write(inPck)

		select {
		case outPck := <-outReader.Read():
			require.NotNil(t, outPck)
			outReader.Receive(outPck)
			require.Error(t, outPck.Payload().(types.Error).Unwrap())
		case <-ctx.Done():
			require.Fail(t, ctx.Err().Error())
		}
	})

	t.Run("PayloadError", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()

		assert := NewAssertNode(func(ctx context.Context, payload any) (bool, error) {
			if val, ok := payload.(int); ok {
				return val == 10, nil
			}
			return false, nil
		})
		defer assert.Close()

		in := port.NewOut()
		in.Link(assert.In(node.PortIn))

		out := port.NewIn()
		assert.Out(node.PortError).Link(out)

		proc := process.New()
		defer proc.Exit(nil)

		inWriter := in.Open(proc)
		outReader := out.Open(proc)

		inPayload, err := types.Marshal([]any{"error", "error"})
		require.NoError(t, err)
		inPck := packet.New(inPayload)

		inWriter.Write(inPck)

		select {
		case outPck := <-outReader.Read():
			require.NotNil(t, outPck)
			outReader.Receive(outPck)
			require.Error(t, outPck.Payload().(types.Error).Unwrap())
		case <-ctx.Done():
			require.Fail(t, ctx.Err().Error())
		}
	})
}
