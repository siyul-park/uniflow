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
	"github.com/siyul-park/uniflow/pkg/primitive"
	"github.com/siyul-park/uniflow/pkg/process"
	"github.com/stretchr/testify/assert"
)

func TestNewLoopNode(t *testing.T) {
	n := NewLoopNode()
	assert.NotNil(t, n)
	assert.Equal(t, 1, n.Batch())
	assert.NoError(t, n.Close())
}

func TestLoopNode_Port(t *testing.T) {
	n := NewLoopNode()
	defer n.Close()

	assert.NotNil(t, n.In(node.PortIn))
	assert.NotNil(t, n.Out(node.PortOut))
	assert.NotNil(t, n.Out(node.PortErr))
	assert.NotNil(t, n.Out(node.PortWithIndex(node.PortOut, 0)))
	assert.NotNil(t, n.Out(node.PortWithIndex(node.PortOut, 1)))
}

func TestLoopNode_SendAndReceive(t *testing.T) {
	t.Run("In -> Out -> In", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.TODO(), time.Second)
		defer cancel()

		n := NewLoopNode()
		defer n.Close()

		in := port.NewOut()
		in.Link(n.In(node.PortIn))

		out := port.NewIn()
		n.Out(node.PortOut).Link(out)

		proc := process.New()
		defer proc.Close()

		inWriter := in.Open(proc)
		outReader := out.Open(proc)

		inPayload := primitive.NewSlice()
		for i := 0; i < 4; i++ {
			inPayload = inPayload.Append(primitive.NewString(faker.UUIDHyphenated()))
		}
		inPck := packet.New(inPayload)

		inWriter.Write(inPck)

		for i := 0; i < inPayload.Len(); i++ {
			select {
			case outPck := <-outReader.Read():
				assert.Equal(t, inPayload.Get(i).Interface(), outPck.Payload().Interface())
				outReader.Receive(outPck)
			case <-ctx.Done():
				assert.Fail(t, "timeout")
			}
		}

		select {
		case <-inWriter.Receive():
		case <-ctx.Done():
			assert.Fail(t, "timeout")
		}
	})

	t.Run("In -> Out0 -> Out1 -> In", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.TODO(), time.Second)
		defer cancel()

		n := NewLoopNode()
		defer n.Close()

		in := port.NewOut()
		in.Link(n.In(node.PortIn))

		out0 := port.NewIn()
		n.Out(node.PortWithIndex(node.PortOut, 0)).Link(out0)

		out1 := port.NewIn()
		n.Out(node.PortWithIndex(node.PortOut, 1)).Link(out1)

		proc := process.New()
		defer proc.Close()

		inWriter := in.Open(proc)
		outReader0 := out0.Open(proc)
		outReader1 := out1.Open(proc)

		inPayload := primitive.NewSlice()
		for i := 0; i < 4; i++ {
			inPayload = inPayload.Append(primitive.NewString(faker.UUIDHyphenated()))
		}
		inPck := packet.New(inPayload)

		inWriter.Write(inPck)

		for i := 0; i < inPayload.Len(); i++ {
			select {
			case outPck := <-outReader0.Read():
				assert.Equal(t, inPayload.Get(i).Interface(), outPck.Payload().Interface())
				outReader0.Receive(outPck)
			case <-ctx.Done():
				assert.Fail(t, "timeout")
			}
		}

		select {
		case outPck := <-outReader1.Read():
			outReader1.Receive(outPck)
		case <-ctx.Done():
			assert.Fail(t, "timeout")
		}

		select {
		case <-inWriter.Receive():
		case <-ctx.Done():
			assert.Fail(t, "timeout")
		}
	})

	t.Run("In -> Out0 -> Error -> In", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.TODO(), time.Second)
		defer cancel()

		n := NewLoopNode()
		defer n.Close()

		in := port.NewOut()
		in.Link(n.In(node.PortIn))

		out0 := port.NewIn()
		n.Out(node.PortWithIndex(node.PortOut, 0)).Link(out0)

		proc := process.New()
		defer proc.Close()

		inWriter := in.Open(proc)
		outReader0 := out0.Open(proc)

		inPayload := primitive.NewSlice()
		for i := 0; i < 4; i++ {
			inPayload = inPayload.Append(primitive.NewString(faker.UUIDHyphenated()))
		}
		inPck := packet.New(inPayload)

		inWriter.Write(inPck)

		select {
		case outPck := <-outReader0.Read():
			backPck := packet.WithError(errors.New(faker.Sentence()), outPck)
			proc.Stack().Add(outPck, backPck)

			outReader0.Receive(backPck)
		case <-ctx.Done():
			assert.Fail(t, "timeout")
		}

		select {
		case <-inWriter.Receive():
		case <-ctx.Done():
			assert.Fail(t, "timeout")
		}
	})

	t.Run("In -> Out0 -> Error -> Out1 -> In", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.TODO(), time.Second)
		defer cancel()

		n := NewLoopNode()
		defer n.Close()

		in := port.NewOut()
		in.Link(n.In(node.PortIn))

		out0 := port.NewIn()
		n.Out(node.PortWithIndex(node.PortOut, 0)).Link(out0)

		out1 := port.NewIn()
		n.Out(node.PortWithIndex(node.PortOut, 1)).Link(out1)

		err := port.NewIn()
		n.Out(node.PortErr).Link(err)

		proc := process.New()
		defer proc.Close()

		inWriter := in.Open(proc)
		outReader0 := out0.Open(proc)
		outReader1 := out1.Open(proc)
		errReader := err.Open(proc)

		inPayload := primitive.NewSlice()
		for i := 0; i < 4; i++ {
			inPayload = inPayload.Append(primitive.NewString(faker.UUIDHyphenated()))
		}
		inPck := packet.New(inPayload)

		inWriter.Write(inPck)

		for i := 0; i < inPayload.Len(); i++ {
			select {
			case outPck := <-outReader0.Read():
				assert.Equal(t, inPayload.Get(i), outPck.Payload())

				backPck := packet.WithError(errors.New(faker.Sentence()), outPck)
				proc.Stack().Add(outPck, backPck)

				outReader0.Receive(backPck)
			case <-ctx.Done():
				assert.Fail(t, "timeout")
			}

			select {
			case outPck := <-errReader.Read():
				backPck := packet.New(primitive.NewString(faker.UUIDHyphenated()))
				proc.Stack().Add(outPck, backPck)

				errReader.Receive(backPck)
			case <-ctx.Done():
				assert.Fail(t, "timeout")
			}
		}

		select {
		case outPck := <-outReader1.Read():
			outReader1.Receive(outPck)
		case <-ctx.Done():
			assert.Fail(t, "timeout")
		}

		select {
		case <-inWriter.Receive():
		case <-ctx.Done():
			assert.Fail(t, "timeout")
		}
	})

	t.Run("batch = 2", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.TODO(), time.Second)
		defer cancel()

		n := NewLoopNode()
		defer n.Close()

		n.SetBatch(2)

		in := port.NewOut()
		in.Link(n.In(node.PortIn))

		out := port.NewIn()
		n.Out(node.PortOut).Link(out)

		proc := process.New()
		defer proc.Close()

		inWriter := in.Open(proc)
		outReader := out.Open(proc)

		inPayload := primitive.NewSlice()
		for i := 0; i < 4; i++ {
			inPayload = inPayload.Append(primitive.NewString(faker.UUIDHyphenated()))
		}
		inPck := packet.New(inPayload)

		inWriter.Write(inPck)

		for i := 0; i < inPayload.Len()/n.Batch(); i++ {
			select {
			case outPck := <-outReader.Read():
				outReader.Receive(outPck)
			case <-ctx.Done():
				assert.Fail(t, "timeout")
			}
		}

		select {
		case <-inWriter.Receive():
		case <-ctx.Done():
			assert.Fail(t, "timeout")
		}
	})
}

func TestLoopNodeCodec_Decode(t *testing.T) {
	codec := NewLoopNodeCodec()

	spec := &LoopNodeSpec{
		Batch: 1,
	}

	n, err := codec.Decode(spec)
	assert.NoError(t, err)
	assert.NotNil(t, n)
	assert.NoError(t, n.Close())
}

func BenchmarkLoopNode_SendAndReceive(b *testing.B) {
	n := NewLoopNode()
	defer n.Close()

	in := port.NewOut()
	in.Link(n.In(node.PortIn))

	out0 := port.NewIn()
	n.Out(node.PortWithIndex(node.PortOut, 0)).Link(out0)

	out1 := port.NewIn()
	n.Out(node.PortWithIndex(node.PortOut, 1)).Link(out1)

	proc := process.New()
	defer proc.Close()

	inWriter := in.Open(proc)
	outReader0 := out0.Open(proc)
	outReader1 := out1.Open(proc)

	inPayload := primitive.NewSlice()
	for i := 0; i < 4; i++ {
		inPayload = inPayload.Append(primitive.NewString(faker.UUIDHyphenated()))
	}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		inPck := packet.New(inPayload)
		inWriter.Write(inPck)

		for i := 0; i < inPayload.Len(); i++ {
			outPck := <-outReader0.Read()
			outReader0.Receive(outPck)
		}

		outPck := <-outReader1.Read()
		outReader1.Receive(outPck)

		<-inWriter.Receive()
	}
}
