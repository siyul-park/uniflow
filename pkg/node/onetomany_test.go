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

func TestNewOneToManyNode(t *testing.T) {
	n := NewOneToManyNode(nil)
	assert.NotNil(t, n)

	assert.NoError(t, n.Close())
}

func TestOneToManyNode_Port(t *testing.T) {
	n := NewOneToManyNode(nil)
	defer n.Close()

	assert.NotNil(t, n.In(PortIn))
	assert.NotNil(t, n.Out(MultiPort(PortOut, 0)))
	assert.NotNil(t, n.Out(PortErr))
}

func TestOneToManyNode_SendAndReceive(t *testing.T) {
	t.Run("In -> None", func(t *testing.T) {
		n := NewOneToManyNode(func(_ *process.Process, inPck *packet.Packet) ([]*packet.Packet, *packet.Packet) {
			return nil, nil
		})
		defer n.Close()

		in := port.NewOut()
		in.Link(n.In(PortIn))

		proc := process.New()
		defer proc.Close()

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

	t.Run("In -> Out0 -> In", func(t *testing.T) {
		n := NewOneToManyNode(func(_ *process.Process, inPck *packet.Packet) ([]*packet.Packet, *packet.Packet) {
			return []*packet.Packet{inPck}, nil
		})
		defer n.Close()

		in := port.NewOut()
		in.Link(n.In(PortIn))

		out0 := port.NewIn()
		n.Out(MultiPort(PortOut, 0)).Link(out0)

		proc := process.New()
		defer proc.Close()
		defer proc.Stack().Close()

		inWriter := in.Open(proc)
		outReader0 := out0.Open(proc)

		inPayload := primitive.NewString(faker.UUIDHyphenated())
		inPck := packet.New(inPayload)

		inWriter.Write(inPck)

		ctx, cancel := context.WithTimeout(context.TODO(), time.Second)
		defer cancel()

		select {
		case outPck := <-outReader0.Read():
			assert.Equal(t, inPayload, outPck.Payload())
			outReader0.Receive(outPck)
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
		n := NewOneToManyNode(func(_ *process.Process, inPck *packet.Packet) ([]*packet.Packet, *packet.Packet) {
			return nil, inPck
		})
		defer n.Close()

		in := port.NewOut()
		in.Link(n.In(PortIn))

		err := port.NewIn()
		n.Out(PortErr).Link(err)

		proc := process.New()
		defer proc.Close()

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

func BenchmarkOneToManyNode_SendAndReceive(b *testing.B) {
	b.Run("In -> Out -> In", func(b *testing.B) {
		n := NewOneToManyNode(func(_ *process.Process, inPck *packet.Packet) ([]*packet.Packet, *packet.Packet) {
			return []*packet.Packet{inPck}, nil
		})
		defer n.Close()

		in := port.NewOut()
		in.Link(n.In(PortIn))

		out0 := port.NewIn()
		n.Out(MultiPort(PortOut, 0)).Link(out0)

		proc := process.New()
		defer proc.Close()

		inWriter := in.Open(proc)
		outReader0 := out0.Open(proc)

		b.ResetTimer()

		for i := 0; i < b.N; i++ {
			inPayload := primitive.NewString(faker.UUIDHyphenated())
			inPck := packet.New(inPayload)

			inWriter.Write(inPck)
			outPck := <-outReader0.Read()
			outReader0.Receive(outPck)
			<-inWriter.Receive()
		}
	})
}
