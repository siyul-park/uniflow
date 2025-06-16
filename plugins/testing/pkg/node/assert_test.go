package node

import (
	"context"
	"testing"
	"time"

	"github.com/go-faker/faker/v4"
	"github.com/gofrs/uuid"
	"github.com/pkg/errors"
	"github.com/siyul-park/uniflow/pkg/language/json"
	"github.com/siyul-park/uniflow/pkg/language/text"
	"github.com/siyul-park/uniflow/pkg/meta"
	"github.com/siyul-park/uniflow/pkg/node"
	"github.com/siyul-park/uniflow/pkg/packet"
	"github.com/siyul-park/uniflow/pkg/port"
	"github.com/siyul-park/uniflow/pkg/process"
	"github.com/siyul-park/uniflow/pkg/runtime"
	"github.com/siyul-park/uniflow/pkg/spec"
	"github.com/siyul-park/uniflow/pkg/symbol"
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
		s := &AssertNodeSpec{
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

		n, err := codec.Compile(s)
		require.NoError(t, err)
		require.NotNil(t, n)
		require.NoError(t, n.Close())
	})

	t.Run("CompileError", func(t *testing.T) {
		s := &AssertNodeSpec{
			Meta: spec.Meta{
				ID:        uuid.Must(uuid.NewV7()),
				Kind:      faker.UUIDHyphenated(),
				Namespace: meta.DefaultNamespace,
				Name:      faker.UUIDHyphenated(),
			},
			Expect: "{ error }",
		}

		n, err := codec.Compile(s)
		require.Error(t, err)
		require.Nil(t, n)
	})
}

func TestAssertNodeCodec_Target(t *testing.T) {
	compiler := text.NewCompiler()
	agent := runtime.NewAgent()
	defer agent.Close()

	codec := NewAssertNodeCodec(compiler, agent)
	require.NotNil(t, codec)

	proc := process.New()
	defer proc.Exit(nil)

	t.Run("FindByID", func(t *testing.T) {
		id := uuid.Must(uuid.NewV7())
		n := &symbol.Symbol{
			Spec: &spec.Meta{
				ID:        id,
				Kind:      faker.UUIDHyphenated(),
				Namespace: meta.DefaultNamespace,
				Name:      faker.UUIDHyphenated(),
			},
			Node: node.NewOneToOneNode(func(_ *process.Process, inPck *packet.Packet) (*packet.Packet, *packet.Packet) {
				return inPck, nil
			}),
		}
		defer n.Close()

		in := port.NewOut()
		defer in.Close()
		out := port.NewIn()
		defer out.Close()

		in.Link(n.In(node.PortIn))
		n.Out(node.PortOut).Link(out)

		agent.Load(n)
		defer agent.Unload(n)

		target := codec.Target(meta.DefaultNamespace, &spec.Port{
			ID:   id,
			Port: node.PortIn,
		})

		inWriter := in.Open(proc)
		outReader := out.Open(proc)

		payload := types.NewSlice(types.NewInt(10), types.NewInt(-1))
		pck := packet.New(payload)
		inWriter.Write(pck)

		outPck := <-outReader.Read()
		outReader.Receive(outPck)
		<-inWriter.Receive()

		result, _, err := target(proc, nil, -1)
		require.NoError(t, err)
		require.Equal(t, types.InterfaceOf(payload), result)
	})

	t.Run("FindByName", func(t *testing.T) {
		name := faker.UUIDHyphenated()
		n := &symbol.Symbol{
			Spec: &spec.Meta{
				ID:        uuid.Must(uuid.NewV7()),
				Kind:      faker.UUIDHyphenated(),
				Namespace: meta.DefaultNamespace,
				Name:      name,
			},
			Node: node.NewOneToOneNode(func(_ *process.Process, inPck *packet.Packet) (*packet.Packet, *packet.Packet) {
				return inPck, nil
			}),
		}
		defer n.Close()

		in := port.NewOut()
		defer in.Close()
		out := port.NewIn()
		defer out.Close()

		in.Link(n.In(node.PortIn))
		n.Out(node.PortOut).Link(out)

		agent.Load(n)
		defer agent.Unload(n)

		target := codec.Target(meta.DefaultNamespace, &spec.Port{
			Name: name,
			Port: node.PortOut,
		})

		inWriter := in.Open(proc)
		outReader := out.Open(proc)

		payload := types.NewSlice(types.NewInt(10), types.NewInt(-1))
		pck := packet.New(payload)
		inWriter.Write(pck)

		outPck := <-outReader.Read()
		outReader.Receive(outPck)
		<-inWriter.Receive()

		result, _, err := target(proc, nil, 0)
		require.NoError(t, err)
		require.Equal(t, types.InterfaceOf(payload), result)
	})

	t.Run("NotFound", func(t *testing.T) {
		target := codec.Target(meta.DefaultNamespace, &spec.Port{
			ID:   uuid.Must(uuid.NewV7()),
			Port: node.PortIn,
		})

		result, _, err := target(proc, nil, 0)
		require.ErrorIs(t, err, ErrAssertFail)
		require.Nil(t, result)
	})
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

		inPayload := types.NewSlice(types.NewInt(10), types.NewInt(-1))
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
			require.ErrorIs(t, outPck.Payload().(types.Error), ErrAssertFail)
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

		inPayload := types.NewSlice(types.NewInt(10), types.NewInt(-1))
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

		inPayload := types.NewSlice(types.NewInt(10), types.NewInt(-1))
		inPck := packet.New(inPayload)

		inWriter.Write(inPck)

		select {
		case outPck := <-outReader.Read():
			require.NotNil(t, outPck)
			outReader.Receive(outPck)
			require.ErrorIs(t, outPck.Payload().(types.Error), ErrAssertFail)
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

		inPayload := types.NewSlice(types.NewInt(10), types.NewInt(-1))
		inPck := packet.New(inPayload)

		inWriter.Write(inPck)

		select {
		case outPck := <-outReader.Read():
			require.NotNil(t, outPck)
			outReader.Receive(outPck)
			require.Error(t, outPck.Payload().(types.Error))
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
			require.Error(t, outPck.Payload().(types.Error))
		case <-ctx.Done():
			require.Fail(t, ctx.Err().Error())
		}
	})
}
