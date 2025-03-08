package control

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/go-faker/faker/v4"
	"github.com/siyul-park/uniflow/pkg/node"
	"github.com/siyul-park/uniflow/pkg/packet"
	"github.com/siyul-park/uniflow/pkg/port"
	"github.com/siyul-park/uniflow/pkg/process"
	"github.com/siyul-park/uniflow/pkg/types"
	"github.com/stretchr/testify/require"
)

func TestPipeNodeCodec_Compile(t *testing.T) {
	codec := NewPipeNodeCodec()

	spec := &PipeNodeSpec{}

	n, err := codec.Compile(spec)
	require.NoError(t, err)
	require.NotNil(t, n)
	require.NoError(t, n.Close())
}

func TestNewPipeNode(t *testing.T) {
	n := NewPipeNode()
	require.NotNil(t, n)
	require.NoError(t, n.Close())
}

func TestPipeNode_Port(t *testing.T) {
	n := NewPipeNode()
	defer n.Close()

	require.NotNil(t, n.In(node.PortIn))
	require.NotNil(t, n.Out(node.PortOut))
	require.NotNil(t, n.Out(node.PortError))
	require.NotNil(t, n.Out(node.PortWithIndex(node.PortOut, 0)))
	require.NotNil(t, n.Out(node.PortWithIndex(node.PortOut, 1)))
}

func TestPipeNode_SendAndReceive(t *testing.T) {
	t.Run("SingleInputToMultipleOutputs", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.TODO(), 60*time.Second)
		defer cancel()

		n1 := node.NewOneToOneNode(func(_ *process.Process, inPck *packet.Packet) (*packet.Packet, *packet.Packet) {
			return inPck, nil
		})
		defer n1.Close()

		n2 := NewPipeNode()
		defer n2.Close()

		n2.Out(node.PortWithIndex(node.PortOut, 0)).Link(n1.In(node.PortIn))

		in := port.NewOut()
		in.Link(n2.In(node.PortIn))

		out1 := port.NewIn()
		n2.Out(node.PortWithIndex(node.PortOut, 1)).Link(out1)

		proc := process.New()
		defer proc.Exit(nil)

		inWriter := in.Open(proc)
		outReader1 := out1.Open(proc)

		inPayload := types.NewString(faker.UUIDHyphenated())
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
			require.Equal(t, inPayload, backPck.Payload())
		case <-ctx.Done():
			require.Fail(t, ctx.Err().Error())
		}
	})

	t.Run("SingleInputToSingleError", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.TODO(), time.Second)
		defer cancel()

		n1 := node.NewOneToOneNode(func(_ *process.Process, inPck *packet.Packet) (*packet.Packet, *packet.Packet) {
			return nil, packet.New(types.NewError(errors.New(faker.Sentence())))
		})
		defer n1.Close()

		n2 := NewPipeNode()
		defer n2.Close()

		n2.Out(node.PortWithIndex(node.PortOut, 0)).Link(n1.In(node.PortIn))

		in := port.NewOut()
		in.Link(n2.In(node.PortIn))

		err := port.NewIn()
		n2.Out(node.PortError).Link(err)

		proc := process.New()
		defer proc.Exit(nil)

		inWriter := in.Open(proc)
		errReader := err.Open(proc)

		inPayload := types.NewString(faker.UUIDHyphenated())
		inPck := packet.New(inPayload)

		inWriter.Write(inPck)

		select {
		case outPck := <-errReader.Read():
			require.NotNil(t, outPck)
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

func BenchmarkPipeNode_SendAndReceive(b *testing.B) {
	n1 := node.NewOneToOneNode(func(_ *process.Process, inPck *packet.Packet) (*packet.Packet, *packet.Packet) {
		return inPck, nil
	})
	defer n1.Close()

	n2 := NewPipeNode()
	defer n2.Close()

	n2.Out(node.PortWithIndex(node.PortOut, 0)).Link(n1.In(node.PortIn))

	in := port.NewOut()
	in.Link(n2.In(node.PortIn))

	out1 := port.NewIn()
	n2.Out(node.PortWithIndex(node.PortOut, 1)).Link(out1)

	proc := process.New()
	defer proc.Exit(nil)

	inWriter := in.Open(proc)
	outReader1 := out1.Open(proc)

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		inPayload := types.NewString(faker.UUIDHyphenated())
		inPck := packet.New(inPayload)

		inWriter.Write(inPck)

		outPck := <-outReader1.Read()
		outReader1.Receive(outPck)

		<-inWriter.Receive()
	}
}
