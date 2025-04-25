package node

import (
	"context"
	"testing"
	"time"

	"github.com/go-faker/faker/v4"
	"github.com/stretchr/testify/require"

	"github.com/siyul-park/uniflow/pkg/packet"
	"github.com/siyul-park/uniflow/pkg/port"
	"github.com/siyul-park/uniflow/pkg/process"
	"github.com/siyul-park/uniflow/pkg/types"
)

func TestNewOneToManyNode(t *testing.T) {
	n := NewOneToManyNode(nil)
	require.NotNil(t, n)
	require.NoError(t, n.Close())
}

func TestOneToManyNode_Port(t *testing.T) {
	n := NewOneToManyNode(nil)
	defer n.Close()

	require.NotNil(t, n.In(PortIn))
	require.NotNil(t, n.Out(PortWithIndex(PortOut, 0)))
	require.NotNil(t, n.Out(PortError))
}

func TestOneToManyNode_SendAndReceive(t *testing.T) {
	t.Run("SingleInputToNoOutput", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.TODO(), time.Second)
		defer cancel()

		n := NewOneToManyNode(func(_ *process.Process, inPck *packet.Packet) ([]*packet.Packet, *packet.Packet) {
			return nil, nil
		})
		defer n.Close()

		in := port.NewOut()
		in.Link(n.In(PortIn))

		proc := process.New()
		defer proc.Exit(nil)

		inWriter := in.Open(proc)

		inPayload := types.NewString(faker.UUIDHyphenated())
		inPck := packet.New(inPayload)

		inWriter.Write(inPck)

		select {
		case <-inWriter.Receive():
		case <-ctx.Done():
			require.Fail(t, ctx.Err().Error())
		}
	})

	t.Run("SingleInputToSingleOutput", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.TODO(), time.Second)
		defer cancel()

		n := NewOneToManyNode(func(_ *process.Process, inPck *packet.Packet) ([]*packet.Packet, *packet.Packet) {
			return []*packet.Packet{inPck}, nil
		})
		defer n.Close()

		in := port.NewOut()
		in.Link(n.In(PortIn))

		out0 := port.NewIn()
		n.Out(PortWithIndex(PortOut, 0)).Link(out0)

		proc := process.New()
		defer proc.Exit(nil)

		inWriter := in.Open(proc)
		outReader0 := out0.Open(proc)

		inPayload := types.NewString(faker.UUIDHyphenated())
		inPck := packet.New(inPayload)

		inWriter.Write(inPck)

		select {
		case outPck := <-outReader0.Read():
			require.Equal(t, inPayload, outPck.Payload())
			outReader0.Receive(outPck)
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

	t.Run("SingleInputToMultipleOutputs", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.TODO(), time.Second*100)
		defer cancel()

		n := NewOneToManyNode(func(_ *process.Process, inPck *packet.Packet) ([]*packet.Packet, *packet.Packet) {
			return []*packet.Packet{inPck, inPck}, nil
		})
		defer n.Close()

		in := port.NewOut()
		in.Link(n.In(PortIn))

		out0 := port.NewIn()
		n.Out(PortWithIndex(PortOut, 0)).Link(out0)

		out1 := port.NewIn()
		n.Out(PortWithIndex(PortOut, 1)).Link(out1)

		proc := process.New()
		defer proc.Exit(nil)

		inWriter := in.Open(proc)
		outReader0 := out0.Open(proc)
		outReader1 := out1.Open(proc)

		inPayload := types.NewString(faker.UUIDHyphenated())
		inPck := packet.New(inPayload)

		inWriter.Write(inPck)

		select {
		case outPck := <-outReader0.Read():
			require.Equal(t, inPayload, outPck.Payload())
			outReader0.Receive(outPck)
		case <-ctx.Done():
			require.Fail(t, ctx.Err().Error())
		}

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

		n := NewOneToManyNode(func(_ *process.Process, inPck *packet.Packet) ([]*packet.Packet, *packet.Packet) {
			return nil, inPck
		})
		defer n.Close()

		in := port.NewOut()
		in.Link(n.In(PortIn))

		err := port.NewIn()
		n.Out(PortError).Link(err)

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

func BenchmarkOneToManyNode_SendAndReceive(b *testing.B) {
	n := NewOneToManyNode(func(_ *process.Process, inPck *packet.Packet) ([]*packet.Packet, *packet.Packet) {
		return []*packet.Packet{inPck}, nil
	})
	defer n.Close()

	in := port.NewOut()
	in.Link(n.In(PortIn))

	out0 := port.NewIn()
	n.Out(PortWithIndex(PortOut, 0)).Link(out0)

	proc := process.New()
	defer proc.Exit(nil)

	inWriter := in.Open(proc)
	outReader0 := out0.Open(proc)

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		inPayload := types.NewString(faker.UUIDHyphenated())
		inPck := packet.New(inPayload)

		inWriter.Write(inPck)

		outPck := <-outReader0.Read()
		outReader0.Receive(outPck)

		<-inWriter.Receive()
	}
}
