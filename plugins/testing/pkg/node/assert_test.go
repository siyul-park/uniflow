package node

import (
	"context"
	"testing"
	"time"

	"github.com/pkg/errors"
	"github.com/siyul-park/uniflow/pkg/language/text"
	"github.com/siyul-park/uniflow/pkg/node"
	"github.com/siyul-park/uniflow/pkg/packet"
	"github.com/siyul-park/uniflow/pkg/port"
	"github.com/siyul-park/uniflow/pkg/process"
	"github.com/siyul-park/uniflow/pkg/runtime"
	"github.com/siyul-park/uniflow/pkg/types"
	"github.com/stretchr/testify/require"
)

func TestNewAssertNodeCodec(t *testing.T) {
	compiler := text.NewCompiler()
	agent := runtime.NewAgent()
	defer agent.Close()

	codec := NewAssertNodeCodec(compiler, agent)
	require.NotNil(t, codec)
}

func TestAssertNode_Port(t *testing.T) {
	n := NewAssertNode(nil)
	defer n.Close()

	require.NotNil(t, n.In(node.PortIn))
	require.NotNil(t, n.Out(node.PortOut))
}

func TestAssertNode_SendAndReceive(t *testing.T) {
	t.Run("DirectAssert", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.TODO(), time.Second)
		defer cancel()

		expect := func(_ context.Context, payload any) (bool, error) {
			return payload == 10, nil
		}

		n := NewAssertNode(expect)
		defer n.Close()

		in := port.NewOut()
		in.Link(n.In(node.PortIn))

		out := port.NewIn()
		n.Out(node.PortOut).Link(out)

		proc := process.New()
		defer proc.Exit(nil)

		inWriter := in.Open(proc)
		outReader := out.Open(proc)

		inPayload := types.NewSlice(types.NewInt(10), types.NewInt(-1))
		inPck := packet.New(inPayload)

		inWriter.Write(inPck)

		select {
		case outPck := <-outReader.Read():
			outPayload, ok := outPck.Payload().(types.Slice)
			require.True(t, ok)
			require.Equal(t, 2, outPayload.Len())
			require.Equal(t, types.NewInt(10), outPayload.Get(0))
			require.Equal(t, types.NewInt(-1), outPayload.Get(1))
			outReader.Receive(outPck)
		case <-ctx.Done():
			require.Fail(t, ctx.Err().Error())
		}

		select {
		case backPck := <-inWriter.Receive():
			require.NotNil(t, backPck)
		case <-ctx.Done():
			require.Fail(t, ctx.Err().Error())
		}
	})

	t.Run("AssertFailed", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.TODO(), time.Second)
		defer cancel()

		expect := func(_ context.Context, payload any) (bool, error) {
			return payload == 10, nil
		}

		n := NewAssertNode(expect)
		defer n.Close()

		in := port.NewOut()
		in.Link(n.In(node.PortIn))

		err := port.NewIn()
		n.Out(node.PortError).Link(err)

		proc := process.New()
		defer proc.Exit(nil)

		inWriter := in.Open(proc)
		errReader := err.Open(proc)

		inPayload := types.NewSlice(types.NewInt(5), types.NewInt(-1))
		inPck := packet.New(inPayload)

		inWriter.Write(inPck)

		select {
		case errPck := <-errReader.Read():
			errPayload, ok := errPck.Payload().(types.Error)
			require.True(t, ok)
			require.ErrorIs(t, errPayload.Unwrap(), ErrAssertFail)
			errReader.Receive(errPck)
		case <-ctx.Done():
			require.Fail(t, ctx.Err().Error())
		}

		select {
		case backPck := <-inWriter.Receive():
			require.NotNil(t, backPck)
		case <-ctx.Done():
			require.Fail(t, ctx.Err().Error())
		}
	})

	t.Run("TargetAssert", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.TODO(), time.Second)
		defer cancel()

		expect := func(_ context.Context, payload any) (bool, error) {
			return payload == 42, nil
		}

		n := NewAssertNode(expect)
		defer n.Close()

		n.SetTarget(func(proc *process.Process, payload any, index int) (any, int, error) {
			return 42, 0, nil
		})

		in := port.NewOut()
		in.Link(n.In(node.PortIn))

		out := port.NewIn()
		n.Out(node.PortOut).Link(out)

		proc := process.New()
		defer proc.Exit(nil)

		inWriter := in.Open(proc)
		outReader := out.Open(proc)

		inPayload := types.NewSlice(types.NewInt(10), types.NewInt(-1))
		inPck := packet.New(inPayload)

		inWriter.Write(inPck)

		select {
		case outPck := <-outReader.Read():
			outPayload, ok := outPck.Payload().(types.Slice)
			require.True(t, ok)
			require.Equal(t, 2, outPayload.Len())
			require.Equal(t, types.NewInt(42), outPayload.Get(0))
			require.Equal(t, types.NewInt(0), outPayload.Get(1))
			outReader.Receive(outPck)
		case <-ctx.Done():
			require.Fail(t, ctx.Err().Error())
		}

		select {
		case backPck := <-inWriter.Receive():
			require.NotNil(t, backPck)
		case <-ctx.Done():
			require.Fail(t, ctx.Err().Error())
		}
	})

	t.Run("TargetNotFound", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.TODO(), time.Second)
		defer cancel()

		expect := func(_ context.Context, payload any) (bool, error) {
			return payload == 42, nil
		}

		n := NewAssertNode(expect)
		defer n.Close()

		n.SetTarget(func(proc *process.Process, payload any, index int) (any, int, error) {
			return nil, -1, errors.WithStack(ErrAssertFail)
		})

		in := port.NewOut()
		in.Link(n.In(node.PortIn))

		err := port.NewIn()
		n.Out(node.PortError).Link(err)

		proc := process.New()
		defer proc.Exit(nil)

		inWriter := in.Open(proc)
		errReader := err.Open(proc)

		inPayload := types.NewSlice(types.NewInt(10), types.NewInt(-1))
		inPck := packet.New(inPayload)

		inWriter.Write(inPck)

		select {
		case errPck := <-errReader.Read():
			errPayload, ok := errPck.Payload().(types.Error)
			require.True(t, ok)
			require.ErrorIs(t, errPayload.Unwrap(), ErrAssertFail)
			errReader.Receive(errPck)
		case <-ctx.Done():
			require.Fail(t, ctx.Err().Error())
		}

		select {
		case backPck := <-inWriter.Receive():
			require.NotNil(t, backPck)
		case <-ctx.Done():
			require.Fail(t, ctx.Err().Error())
		}
	})
}
