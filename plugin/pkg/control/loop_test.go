package control

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/go-faker/faker/v4"
	"github.com/siyul-park/uniflow/pkg/node"
	"github.com/siyul-park/uniflow/pkg/object"
	"github.com/siyul-park/uniflow/pkg/packet"
	"github.com/siyul-park/uniflow/pkg/port"
	"github.com/siyul-park/uniflow/pkg/process"
	"github.com/stretchr/testify/assert"
)

func TestNewLoopNode(t *testing.T) {
	n := NewLoopNode()
	assert.NotNil(t, n)
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
	t.Run("SingleInputToSingleOutput", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.TODO(), time.Second)
		defer cancel()

		n := NewLoopNode()
		defer n.Close()

		in := port.NewOut()
		in.Link(n.In(node.PortIn))

		out := port.NewIn()
		n.Out(node.PortOut).Link(out)

		proc := process.New()
		defer proc.Exit(nil)

		inWriter := in.Open(proc)
		outReader := out.Open(proc)

		inPayload := object.NewSlice()
		for i := 0; i < 4; i++ {
			inPayload = inPayload.Append(object.NewString(faker.UUIDHyphenated()))
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

	t.Run("SingleInputToMultipleOutputs", func(t *testing.T) {
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
		defer proc.Exit(nil)

		inWriter := in.Open(proc)
		outReader0 := out0.Open(proc)
		outReader1 := out1.Open(proc)

		inPayload := object.NewSlice()
		for i := 0; i < 4; i++ {
			inPayload = inPayload.Append(object.NewString(faker.UUIDHyphenated()))
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

	t.Run("SingleInputToSingleError", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.TODO(), time.Second)
		defer cancel()

		n := NewLoopNode()
		defer n.Close()

		in := port.NewOut()
		in.Link(n.In(node.PortIn))

		out0 := port.NewIn()
		n.Out(node.PortWithIndex(node.PortOut, 0)).Link(out0)

		proc := process.New()
		defer proc.Exit(nil)

		inWriter := in.Open(proc)
		outReader0 := out0.Open(proc)

		inPayload := object.NewSlice()
		for i := 0; i < 4; i++ {
			inPayload = inPayload.Append(object.NewString(faker.UUIDHyphenated()))
		}
		inPck := packet.New(inPayload)

		inWriter.Write(inPck)

		for i := 0; i < inPayload.Len(); i++ {
			select {
			case <-outReader0.Read():
				backPck := packet.WithError(errors.New(faker.Sentence()))
				outReader0.Receive(backPck)
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

	t.Run("SingleInputToSingleErrorAndMultipleOutputs", func(t *testing.T) {
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
		defer proc.Exit(nil)

		inWriter := in.Open(proc)
		outReader0 := out0.Open(proc)
		outReader1 := out1.Open(proc)
		errReader := err.Open(proc)

		inPayload := object.NewSlice()
		for i := 0; i < 4; i++ {
			inPayload = inPayload.Append(object.NewString(faker.UUIDHyphenated()))
		}
		inPck := packet.New(inPayload)

		inWriter.Write(inPck)

		for i := 0; i < inPayload.Len(); i++ {
			select {
			case outPck := <-outReader0.Read():
				assert.Equal(t, inPayload.Get(i), outPck.Payload())

				backPck := packet.WithError(errors.New(faker.Sentence()))
				outReader0.Receive(backPck)
			case <-ctx.Done():
				assert.Fail(t, "timeout")
			}

			select {
			case <-errReader.Read():
				backPck := packet.New(object.NewString(faker.UUIDHyphenated()))
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
}

func TestLoopNodeCodec_Decode(t *testing.T) {
	codec := NewLoopNodeCodec()

	spec := &LoopNodeSpec{}

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
	defer proc.Exit(nil)

	inWriter := in.Open(proc)
	outReader0 := out0.Open(proc)
	outReader1 := out1.Open(proc)

	inPayload := object.NewSlice()
	for i := 0; i < 4; i++ {
		inPayload = inPayload.Append(object.NewString(faker.UUIDHyphenated()))
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
