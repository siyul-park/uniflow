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

func TestNewOneToOneNode(t *testing.T) {
	n := NewOneToOneNode(nil)
	assert.NotNil(t, n)

	assert.NoError(t, n.Close())
}

func TestOneToOneNode_Port(t *testing.T) {
	n := NewOneToOneNode(nil)
	defer n.Close()

	assert.NotNil(t, n.In(PortIO))
	assert.NotNil(t, n.In(PortIn))
	assert.NotNil(t, n.Out(PortOut))
	assert.NotNil(t, n.Out(PortErr))
}

func TestOneToOneNode_SendAndReceive(t *testing.T) {
	t.Run("IO -> IO", func(t *testing.T) {
		n := NewOneToOneNode(func(_ *process.Process, inPck *packet.Packet) (*packet.Packet, *packet.Packet) {
			return inPck, nil
		})
		defer n.Close()

		io := port.NewOut()
		io.Link(n.In(PortIO))

		proc := process.New()
		defer proc.Exit(nil)

		ioWriter := io.Open(proc)

		inPayload := primitive.NewString(faker.UUIDHyphenated())
		inPck := packet.New(inPayload)

		ioWriter.Write(inPck)

		ctx, cancel := context.WithTimeout(context.TODO(), time.Second)
		defer cancel()

		select {
		case outPck := <-ioWriter.Receive():
			assert.Equal(t, inPayload, outPck.Payload())
		case <-ctx.Done():
			assert.Fail(t, "timeout")
		}
	})

	t.Run("IO -> Error -> IO", func(t *testing.T) {
		n := NewOneToOneNode(func(_ *process.Process, inPck *packet.Packet) (*packet.Packet, *packet.Packet) {
			return nil, packet.New(primitive.NewString(faker.UUIDHyphenated()))
		})
		defer n.Close()

		io := port.NewOut()
		io.Link(n.In(PortIO))

		err := port.NewIn()
		n.Out(PortErr).Link(err)

		proc := process.New()
		defer proc.Exit(nil)

		ioWriter := io.Open(proc)
		errReader := err.Open(proc)

		inPayload := primitive.NewString(faker.UUIDHyphenated())
		inPck := packet.New(inPayload)

		ioWriter.Write(inPck)

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
		case backPck := <-ioWriter.Receive():
			assert.NotNil(t, backPck)
		case <-ctx.Done():
			assert.Fail(t, "timeout")
		}
	})

	t.Run("In -> None", func(t *testing.T) {
		n := NewOneToOneNode(func(_ *process.Process, inPck *packet.Packet) (*packet.Packet, *packet.Packet) {
			return nil, nil
		})
		defer n.Close()

		in := port.NewOut()
		in.Link(n.In(PortIn))

		proc := process.New()
		defer proc.Exit(nil)

		inWriter := in.Open(proc)

		inPayload := primitive.NewString(faker.UUIDHyphenated())
		inPck := packet.New(inPayload)

		inWriter.Write(inPck)

		ctx, cancel := context.WithTimeout(context.TODO(), time.Second)
		defer cancel()

		select {
		case <-proc.Stack().Done(inPck):
		case <-ctx.Done():
			assert.Fail(t, "timeout")
		}
	})

	t.Run("In -> Out -> In", func(t *testing.T) {
		n := NewOneToOneNode(func(_ *process.Process, inPck *packet.Packet) (*packet.Packet, *packet.Packet) {
			return inPck, nil
		})
		defer n.Close()

		in := port.NewOut()
		in.Link(n.In(PortIn))

		out := port.NewIn()
		n.Out(PortOut).Link(out)

		proc := process.New()
		defer proc.Exit(nil)

		inWriter := in.Open(proc)
		outReader := out.Open(proc)

		inPayload := primitive.NewString(faker.UUIDHyphenated())
		inPck := packet.New(inPayload)

		inWriter.Write(inPck)

		ctx, cancel := context.WithTimeout(context.TODO(), time.Second)
		defer cancel()

		select {
		case outPck := <-outReader.Read():
			assert.Equal(t, inPayload, outPck.Payload())
			outReader.Receive(outPck)
		case <-ctx.Done():
			assert.Fail(t, "timeout")
		}

		select {
		case backPck := <-inWriter.Receive():
			assert.NotNil(t, backPck)
		case <-ctx.Done():
			assert.Fail(t, "timeout")
		}
	})

	t.Run("In -> Error -> In", func(t *testing.T) {
		n := NewOneToOneNode(func(_ *process.Process, inPck *packet.Packet) (*packet.Packet, *packet.Packet) {
			return nil, packet.New(primitive.NewString(faker.UUIDHyphenated()))
		})
		defer n.Close()

		in := port.NewOut()
		in.Link(n.In(PortIn))

		err := port.NewIn()
		n.Out(PortErr).Link(err)

		proc := process.New()
		defer proc.Exit(nil)

		inWriter := in.Open(proc)
		errReader := err.Open(proc)

		inPayload := primitive.NewString(faker.UUIDHyphenated())
		inPck := packet.New(inPayload)

		inWriter.Write(inPck)

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
		case backPck := <-inWriter.Receive():
			assert.NotNil(t, backPck)
		case <-ctx.Done():
			assert.Fail(t, "timeout")
		}
	})
}

func BenchmarkOneToOneNode_SendAndReceive(b *testing.B) {
	b.Run("IO -> IO", func(b *testing.B) {
		n := NewOneToOneNode(func(_ *process.Process, inPck *packet.Packet) (*packet.Packet, *packet.Packet) {
			return inPck, nil
		})
		defer n.Close()

		io := port.NewOut()
		io.Link(n.In(PortIO))

		proc := process.New()
		defer proc.Exit(nil)

		ioWriter := io.Open(proc)

		b.ResetTimer()

		for i := 0; i < b.N; i++ {
			inPayload := primitive.NewString(faker.UUIDHyphenated())
			inPck := packet.New(inPayload)

			ioWriter.Write(inPck)
			<-ioWriter.Receive()
		}
	})

	b.Run("In -> Out -> In", func(b *testing.B) {
		n := NewOneToOneNode(func(_ *process.Process, inPck *packet.Packet) (*packet.Packet, *packet.Packet) {
			return inPck, nil
		})
		defer n.Close()

		in := port.NewOut()
		in.Link(n.In(PortIn))

		out := port.NewIn()
		n.Out(PortOut).Link(out)

		proc := process.New()
		defer proc.Exit(nil)

		inWriter := in.Open(proc)
		outReader := out.Open(proc)

		b.ResetTimer()

		for i := 0; i < b.N; i++ {
			inPayload := primitive.NewString(faker.UUIDHyphenated())
			inPck := packet.New(inPayload)

			inWriter.Write(inPck)
			outPck := <-outReader.Read()
			outReader.Receive(outPck)
			<-inWriter.Receive()
		}
	})
}
