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
	"github.com/stretchr/testify/assert"
)

func TestForNodeCodec_Compile(t *testing.T) {
	codec := NewForNodeCodec()

	spec := &ForNodeSpec{}

	n, err := codec.Compile(spec)
	assert.NoError(t, err)
	assert.NotNil(t, n)
	assert.NoError(t, n.Close())
}

func TestNewForNode(t *testing.T) {
	n := NewForNode()
	assert.NotNil(t, n)
	assert.NoError(t, n.Close())
}

func TestForNode_Port(t *testing.T) {
	n := NewForNode()
	defer n.Close()

	assert.NotNil(t, n.In(node.PortIn))
	assert.NotNil(t, n.Out(node.PortOut))
	assert.NotNil(t, n.Out(node.PortError))
	assert.NotNil(t, n.Out(node.PortWithIndex(node.PortOut, 0)))
	assert.NotNil(t, n.Out(node.PortWithIndex(node.PortOut, 1)))
}

func TestForNode_SendAndReceive(t *testing.T) {
	t.Run("SingleInputToSingleOutput", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.TODO(), time.Second)
		defer cancel()

		n := NewForNode()
		defer n.Close()

		in := port.NewOut()
		in.Link(n.In(node.PortIn))

		out := port.NewIn()
		n.Out(node.PortOut).Link(out)

		proc := process.New()
		defer proc.Exit(nil)

		inWriter := in.Open(proc)
		outReader := out.Open(proc)

		inPayload := types.NewSlice()
		for i := 0; i < 4; i++ {
			inPayload = inPayload.Append(types.NewString(faker.UUIDHyphenated()))
		}
		inPck := packet.New(inPayload)

		inWriter.Write(inPck)

		for i := 0; i < inPayload.Len(); i++ {
			select {
			case outPck := <-outReader.Read():
				assert.Equal(t, inPayload.Get(i), outPck.Payload())
				outReader.Receive(outPck)
			case <-ctx.Done():
				assert.Fail(t, ctx.Err().Error())
			}
		}

		select {
		case <-inWriter.Receive():
		case <-ctx.Done():
			assert.Fail(t, ctx.Err().Error())
		}
	})

	t.Run("SingleInputToMultipleOutputs", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.TODO(), time.Second)
		defer cancel()

		n := NewForNode()
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

		inPayload := types.NewSlice()
		for i := 0; i < 4; i++ {
			inPayload = inPayload.Append(types.NewString(faker.UUIDHyphenated()))
		}
		inPck := packet.New(inPayload)

		inWriter.Write(inPck)

		for i := 0; i < inPayload.Len(); i++ {
			select {
			case outPck := <-outReader0.Read():
				assert.Equal(t, inPayload.Get(i), outPck.Payload())
				outReader0.Receive(outPck)
			case <-ctx.Done():
				assert.Fail(t, ctx.Err().Error())
			}
		}

		select {
		case outPck := <-outReader1.Read():
			outReader1.Receive(outPck)
		case <-ctx.Done():
			assert.Fail(t, ctx.Err().Error())
		}

		select {
		case <-inWriter.Receive():
		case <-ctx.Done():
			assert.Fail(t, ctx.Err().Error())
		}
	})

	t.Run("SingleInputToSingleError", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.TODO(), time.Second)
		defer cancel()

		n := NewForNode()
		defer n.Close()

		in := port.NewOut()
		in.Link(n.In(node.PortIn))

		out0 := port.NewIn()
		n.Out(node.PortWithIndex(node.PortOut, 0)).Link(out0)

		proc := process.New()
		defer proc.Exit(nil)

		inWriter := in.Open(proc)
		outReader0 := out0.Open(proc)

		inPayload := types.NewSlice()
		for i := 0; i < 4; i++ {
			inPayload = inPayload.Append(types.NewString(faker.UUIDHyphenated()))
		}
		inPck := packet.New(inPayload)

		inWriter.Write(inPck)

		for i := 0; i < inPayload.Len(); i++ {
			select {
			case <-outReader0.Read():
				backPck := packet.New(types.NewError(errors.New(faker.Sentence())))
				outReader0.Receive(backPck)
			case <-ctx.Done():
				assert.Fail(t, ctx.Err().Error())
			}
		}

		select {
		case <-inWriter.Receive():
		case <-ctx.Done():
			assert.Fail(t, ctx.Err().Error())
		}
	})
}

func BenchmarkForNode_SendAndReceive(b *testing.B) {
	n := NewForNode()
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

	inPayload := types.NewSlice()
	for i := 0; i < 4; i++ {
		inPayload = inPayload.Append(types.NewString(faker.UUIDHyphenated()))
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
