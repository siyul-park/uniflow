package control

import (
	"context"
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

func TestNewIterateNode(t *testing.T) {
	n := NewIterateNode()
	assert.NotNil(t, n)

	assert.NoError(t, n.Close())
}

func TestIterateNode_Port(t *testing.T) {
	n := NewIterateNode()
	defer n.Close()

	assert.NotNil(t, n.In(node.PortIn))
	assert.NotNil(t, n.Out(node.PortOut))
	assert.NotNil(t, n.Out(node.PortWithIndex(node.PortOut, 0)))
	assert.NotNil(t, n.Out(node.PortWithIndex(node.PortOut, 1)))
}

func TestIterateNode_SendAndReceive(t *testing.T) {
	t.Run("In -> Out -> In", func(t *testing.T) {
		n := NewIterateNode()
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

		ctx, cancel := context.WithTimeout(context.TODO(), time.Second)
		defer cancel()

		for i := 0; i < inPayload.Len(); i++ {
			select {
			case outPck := <-outReader.Read():
				assert.Equal(t, inPayload.Get(i), outPck.Payload())
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
		n := NewIterateNode()
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

		ctx, cancel := context.WithTimeout(context.TODO(), time.Second)
		defer cancel()

		for i := 0; i < inPayload.Len(); i++ {
			select {
			case outPck := <-outReader0.Read():
				assert.Equal(t, inPayload.Get(i), outPck.Payload())
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
}
