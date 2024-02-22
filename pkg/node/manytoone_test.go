package node

import (
	"context"
	"testing"
	"time"

	"github.com/go-faker/faker/v4"
	"github.com/siyul-park/uniflow/pkg/packet"
	"github.com/siyul-park/uniflow/pkg/port"
	"github.com/siyul-park/uniflow/pkg/primitive"
	"github.com/siyul-park/uniflow/pkg/process"
	"github.com/stretchr/testify/assert"
)

func TestNewManyToOneNode(t *testing.T) {
	n := NewManyToOneNode(nil)
	assert.NotNil(t, n)

	assert.NoError(t, n.Close())
}

func TestManyToOneNode_Port(t *testing.T) {
	n := NewManyToOneNode(nil)
	assert.NotNil(t, n)

	assert.NotNil(t, n.In(MultiPort(PortIn, 0)))
	assert.NotNil(t, n.Out(PortOut))
	assert.NotNil(t, n.Out(PortErr))
}

func TestManyToOneNode_SendAndReceive(t *testing.T) {
	t.Run("In0 -> None", func(t *testing.T) {
		n := NewManyToOneNode(func(_ *process.Process, inPcks []*packet.Packet) (*packet.Packet, *packet.Packet) {
			return nil, nil
		})
		defer n.Close()

		in0 := port.NewOut()
		in0.Link(n.In(MultiPort(PortIn, 0)))

		proc := process.New()
		defer proc.Exit(nil)

		inWriter0 := in0.Open(proc)

		inPayload := primitive.NewString(faker.UUIDHyphenated())
		inPck := packet.New(inPayload)

		inWriter0.Write(inPck)

		ctx, cancel := context.WithTimeout(context.TODO(), time.Second)
		defer cancel()

		select {
		case <-proc.Stack().Done(inPck):
		case <-ctx.Done():
			assert.Fail(t, "timeout")
		}
	})

	t.Run("In0, In1 -> Out -> In0, In1", func(t *testing.T) {
		n := NewManyToOneNode(func(_ *process.Process, inPcks []*packet.Packet) (*packet.Packet, *packet.Packet) {
			for _, inPck := range inPcks {
				if inPck == nil {
					return nil, nil
				}
			}
			return packet.New(primitive.NewString(faker.UUIDHyphenated())), nil
		})
		defer n.Close()

		in0 := port.NewOut()
		in0.Link(n.In(MultiPort(PortIn, 0)))

		in1 := port.NewOut()
		in1.Link(n.In(MultiPort(PortIn, 1)))

		out := port.NewIn()
		n.Out(PortOut).Link(out)

		proc := process.New()
		defer proc.Exit(nil)

		inWriter0 := in0.Open(proc)
		inWriter1 := in1.Open(proc)
		outReader := out.Open(proc)

		inPayload := primitive.NewString(faker.UUIDHyphenated())
		inPck0 := packet.New(inPayload)
		inPck1 := packet.New(inPayload)

		inWriter0.Write(inPck0)
		inWriter1.Write(inPck1)

		ctx, cancel := context.WithTimeout(context.TODO(), time.Second)
		defer cancel()

		select {
		case outPck := <-outReader.Read():
			assert.NotNil(t, outPck)
			outReader.Receive(outPck)
		case <-ctx.Done():
			assert.Fail(t, "timeout")
		}

		select {
		case backPck := <-inWriter0.Receive():
			assert.NotNil(t, backPck)
		case <-ctx.Done():
			assert.Fail(t, "timeout")
		}

		select {
		case backPck := <-inWriter1.Receive():
			assert.NotNil(t, backPck)
		case <-ctx.Done():
			assert.Fail(t, "timeout")
		}
	})

	t.Run("In0, In1 -> Error -> In0, In1", func(t *testing.T) {
		n := NewManyToOneNode(func(_ *process.Process, inPcks []*packet.Packet) (*packet.Packet, *packet.Packet) {
			for _, inPck := range inPcks {
				if inPck == nil {
					return nil, nil
				}
			}
			return nil, packet.New(primitive.NewString(faker.UUIDHyphenated()))
		})
		defer n.Close()

		in0 := port.NewOut()
		in0.Link(n.In(MultiPort(PortIn, 0)))

		in1 := port.NewOut()
		in1.Link(n.In(MultiPort(PortIn, 1)))

		err := port.NewIn()
		n.Out(PortErr).Link(err)

		proc := process.New()
		defer proc.Exit(nil)

		inWriter0 := in0.Open(proc)
		inWriter1 := in1.Open(proc)
		errReader := err.Open(proc)

		inPayload := primitive.NewString(faker.UUIDHyphenated())
		inPck0 := packet.New(inPayload)
		inPck1 := packet.New(inPayload)

		inWriter0.Write(inPck0)
		inWriter1.Write(inPck1)

		ctx, cancel := context.WithTimeout(context.TODO(), time.Second)
		defer cancel()

		select {
		case outPck := <-errReader.Read():
			assert.NotNil(t, outPck)
			errReader.Receive(outPck)
		case <-ctx.Done():
			assert.Fail(t, "timeout")
		}

		select {
		case backPck := <-inWriter0.Receive():
			assert.NotNil(t, backPck)
		case <-ctx.Done():
			assert.Fail(t, "timeout")
		}

		select {
		case backPck := <-inWriter1.Receive():
			assert.NotNil(t, backPck)
		case <-ctx.Done():
			assert.Fail(t, "timeout")
		}
	})
}

func BenchmarkManyToOneNode_SendAndReceive(b *testing.B) {
	b.Run("In0, In1 -> Out -> In0, In1", func(b *testing.B) {
		n := NewManyToOneNode(func(_ *process.Process, inPcks []*packet.Packet) (*packet.Packet, *packet.Packet) {
			for _, inPck := range inPcks {
				if inPck == nil {
					return nil, nil
				}
			}
			return packet.New(primitive.NewString(faker.UUIDHyphenated())), nil
		})
		defer n.Close()

		in0 := port.NewOut()
		in0.Link(n.In(MultiPort(PortIn, 0)))

		in1 := port.NewOut()
		in1.Link(n.In(MultiPort(PortIn, 1)))

		out := port.NewIn()
		n.Out(PortOut).Link(out)

		proc := process.New()
		defer proc.Exit(nil)

		inWriter0 := in0.Open(proc)
		inWriter1 := in1.Open(proc)
		outReader := out.Open(proc)

		b.ResetTimer()

		for i := 0; i < b.N; i++ {
			inPayload := primitive.NewString(faker.UUIDHyphenated())
			inPck0 := packet.New(inPayload)
			inPck1 := packet.New(inPayload)

			inWriter0.Write(inPck0)
			inWriter1.Write(inPck1)
			outPck := <-outReader.Read()
			outReader.Receive(outPck)
			<-inWriter0.Receive()
			<-inWriter1.Receive()
		}
	})
}
