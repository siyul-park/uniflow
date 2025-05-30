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

func TestNewManyToOneNode(t *testing.T) {
	n := NewManyToOneNode(nil)
	require.NotNil(t, n)
	require.NoError(t, n.Close())
}

func TestManyToOneNode_Port(t *testing.T) {
	n := NewManyToOneNode(nil)
	defer n.Close()

	require.NotNil(t, n.In(PortWithIndex(PortIn, 0)))
	require.NotNil(t, n.Out(PortOut))
	require.NotNil(t, n.Out(PortError))
}

func TestManyToOneNode_SendAndReceive(t *testing.T) {
	t.Run("SingleInputToNoOutput", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.TODO(), time.Second)
		defer cancel()

		n := NewManyToOneNode(func(_ *process.Process, inPcks []*packet.Packet) (*packet.Packet, *packet.Packet) {
			return nil, nil
		})
		defer n.Close()

		in0 := port.NewOut()
		in0.Link(n.In(PortWithIndex(PortIn, 0)))

		proc := process.New()
		defer proc.Exit(nil)

		inWriter0 := in0.Open(proc)

		inPayload := types.NewString(faker.UUIDHyphenated())
		inPck := packet.New(inPayload)

		inWriter0.Write(inPck)

		select {
		case <-inWriter0.Receive():
		case <-ctx.Done():
			require.Fail(t, ctx.Err().Error())
		}
	})

	t.Run("MultipleInputsToSingleOutput", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.TODO(), time.Second)
		defer cancel()

		n := NewManyToOneNode(func(_ *process.Process, inPcks []*packet.Packet) (*packet.Packet, *packet.Packet) {
			for _, inPck := range inPcks {
				if inPck == nil {
					return nil, nil
				}
			}
			return packet.New(types.NewString(faker.UUIDHyphenated())), nil
		})
		defer n.Close()

		in0 := port.NewOut()
		in0.Link(n.In(PortWithIndex(PortIn, 0)))

		in1 := port.NewOut()
		in1.Link(n.In(PortWithIndex(PortIn, 1)))

		out := port.NewIn()
		n.Out(PortOut).Link(out)

		proc := process.New()
		defer proc.Exit(nil)

		inWriter0 := in0.Open(proc)
		inWriter1 := in1.Open(proc)
		outReader := out.Open(proc)

		inPayload := types.NewString(faker.UUIDHyphenated())
		inPck0 := packet.New(inPayload)
		inPck1 := packet.New(inPayload)

		inWriter0.Write(inPck0)
		inWriter1.Write(inPck1)

		select {
		case outPck := <-outReader.Read():
			require.NotNil(t, outPck)
			outReader.Receive(outPck)
		case <-ctx.Done():
			require.Fail(t, ctx.Err().Error())
		}

		select {
		case backPck := <-inWriter0.Receive():
			require.NotNil(t, backPck)
		case <-ctx.Done():
			require.Fail(t, ctx.Err().Error())
		}

		select {
		case backPck := <-inWriter1.Receive():
			require.NotNil(t, backPck)
		case <-ctx.Done():
			require.Fail(t, ctx.Err().Error())
		}
	})

	t.Run("MultipleInputsToSingleError", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.TODO(), time.Second)
		defer cancel()

		n := NewManyToOneNode(func(_ *process.Process, inPcks []*packet.Packet) (*packet.Packet, *packet.Packet) {
			for _, inPck := range inPcks {
				if inPck == nil {
					return nil, nil
				}
			}
			return nil, packet.New(types.NewString(faker.UUIDHyphenated()))
		})
		defer n.Close()

		in0 := port.NewOut()
		in0.Link(n.In(PortWithIndex(PortIn, 0)))

		in1 := port.NewOut()
		in1.Link(n.In(PortWithIndex(PortIn, 1)))

		err := port.NewIn()
		n.Out(PortError).Link(err)

		proc := process.New()
		defer proc.Exit(nil)

		inWriter0 := in0.Open(proc)
		inWriter1 := in1.Open(proc)
		errReader := err.Open(proc)

		inPayload := types.NewString(faker.UUIDHyphenated())
		inPck0 := packet.New(inPayload)
		inPck1 := packet.New(inPayload)

		inWriter0.Write(inPck0)
		inWriter1.Write(inPck1)

		select {
		case outPck := <-errReader.Read():
			require.NotNil(t, outPck)
			errReader.Receive(outPck)
		case <-ctx.Done():
			require.Fail(t, ctx.Err().Error())
		}

		select {
		case backPck := <-inWriter0.Receive():
			require.NotNil(t, backPck)
		case <-ctx.Done():
			require.Fail(t, ctx.Err().Error())
		}

		select {
		case backPck := <-inWriter1.Receive():
			require.NotNil(t, backPck)
		case <-ctx.Done():
			require.Fail(t, ctx.Err().Error())
		}
	})
}

func BenchmarkManyToOneNode_SendAndReceive(b *testing.B) {
	n := NewManyToOneNode(func(_ *process.Process, inPcks []*packet.Packet) (*packet.Packet, *packet.Packet) {
		for _, inPck := range inPcks {
			if inPck == nil {
				return nil, nil
			}
		}
		return packet.New(types.NewString(faker.UUIDHyphenated())), nil
	})
	defer n.Close()

	in0 := port.NewOut()
	in0.Link(n.In(PortWithIndex(PortIn, 0)))

	in1 := port.NewOut()
	in1.Link(n.In(PortWithIndex(PortIn, 1)))

	out := port.NewIn()
	n.Out(PortOut).Link(out)

	proc := process.New()
	defer proc.Exit(nil)

	inWriter0 := in0.Open(proc)
	inWriter1 := in1.Open(proc)
	outReader := out.Open(proc)

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		inPayload := types.NewString(faker.UUIDHyphenated())
		inPck0 := packet.New(inPayload)
		inPck1 := packet.New(inPayload)

		inWriter0.Write(inPck0)
		inWriter1.Write(inPck1)

		outPck := <-outReader.Read()
		outReader.Receive(outPck)

		<-inWriter0.Receive()
		<-inWriter1.Receive()
	}
}
