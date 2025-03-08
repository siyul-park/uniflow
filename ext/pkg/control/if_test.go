package control

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/go-faker/faker/v4"
	"github.com/siyul-park/uniflow/ext/pkg/language/text"
	"github.com/siyul-park/uniflow/pkg/node"
	"github.com/siyul-park/uniflow/pkg/packet"
	"github.com/siyul-park/uniflow/pkg/port"
	"github.com/siyul-park/uniflow/pkg/process"
	"github.com/siyul-park/uniflow/pkg/types"
	"github.com/stretchr/testify/require"
)

func TestIfNodeCodec_Compile(t *testing.T) {
	codec := NewIfNodeCodec(text.NewCompiler())

	spec := &IfNodeSpec{
		When: "",
	}

	n, err := codec.Compile(spec)
	require.NoError(t, err)
	require.NotNil(t, n)
	require.NoError(t, n.Close())
}

func TestNewIfNode(t *testing.T) {
	n := NewIfNode(nil)
	require.NotNil(t, n)
	require.NoError(t, n.Close())
}

func TestIfNode_SendAndReceive(t *testing.T) {
	t.Run("SingleInputToFirstOutput", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.TODO(), time.Second)
		defer cancel()

		n := NewIfNode(func(_ context.Context, _ any) (bool, error) {
			return true, nil
		})
		defer n.Close()

		in := port.NewOut()
		in.Link(n.In(node.PortIn))

		out := port.NewIn()
		n.Out(node.PortOut).Link(out)

		proc := process.New()
		defer proc.Exit(nil)

		inWriter := in.Open(proc)
		outReader := out.Open(proc)

		inPayload := types.NewMap(types.NewString("foo"), types.NewString("bar"))
		inPck := packet.New(inPayload)

		inWriter.Write(inPck)

		select {
		case outPck := <-outReader.Read():
			require.Equal(t, inPayload, outPck.Payload())
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

	t.Run("SingleInputToSecondOutput", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.TODO(), time.Second)
		defer cancel()

		n := NewIfNode(func(_ context.Context, _ any) (bool, error) {
			return false, nil
		})
		defer n.Close()

		in := port.NewOut()
		in.Link(n.In(node.PortIn))

		out1 := port.NewIn()
		n.Out(node.PortWithIndex(node.PortOut, 1)).Link(out1)

		proc := process.New()
		defer proc.Exit(nil)

		inWriter := in.Open(proc)
		outReader1 := out1.Open(proc)

		inPayload := types.NewMap(types.NewString("foo"), types.NewString("foo"))
		inPck := packet.New(inPayload)

		inWriter.Write(inPck)

		select {
		case outPck := <-outReader1.Read():
			require.Equal(t, inPayload, outPck.Payload())
			outReader1.Receive(outPck)
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

	t.Run("SingleInputToSingleError", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.TODO(), time.Second)
		defer cancel()

		n := NewIfNode(func(_ context.Context, _ any) (bool, error) {
			return false, errors.New(faker.Sentence())
		})
		defer n.Close()

		in := port.NewOut()
		in.Link(n.In(node.PortIn))

		err := port.NewIn()
		n.Out(node.PortError).Link(err)

		proc := process.New()
		defer proc.Exit(nil)

		inWriter := in.Open(proc)
		errReader := err.Open(proc)

		inPayload := types.NewMap(types.NewString("foo"), types.NewString("bar"))
		inPck := packet.New(inPayload)

		inWriter.Write(inPck)

		select {
		case outPck := <-errReader.Read():
			errReader.Receive(outPck)
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

func BenchmarkIfNode_SendAndReceive(b *testing.B) {
	n := NewIfNode(func(_ context.Context, _ any) (bool, error) {
		return true, nil
	})
	defer n.Close()

	in := port.NewOut()
	in.Link(n.In(node.PortIn))

	out0 := port.NewIn()
	n.Out(node.PortWithIndex(node.PortOut, 0)).Link(out0)

	proc := process.New()
	defer proc.Exit(nil)

	inWriter := in.Open(proc)
	outReader0 := out0.Open(proc)

	inPayload := types.NewMap(types.NewString("foo"), types.NewString("bar"))
	inPck := packet.New(inPayload)

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		inWriter.Write(inPck)

		outPck := <-outReader0.Read()
		outReader0.Receive(outPck)

		<-inWriter.Receive()
	}
}
