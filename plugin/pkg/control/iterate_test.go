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
	n := NewIterateNode(0)
	assert.NotNil(t, n)

	assert.NoError(t, n.Close())
}

func TestIterateNode_Port(t *testing.T) {
	n := NewIterateNode(0)
	defer n.Close()

	p := n.Port(node.PortIn)
	assert.NotNil(t, p)

	p = n.Port(node.PortOut)
	assert.NotNil(t, p)

	p = n.Port(node.MultiPort(node.PortOut, 0))
	assert.NotNil(t, p)

	p = n.Port(node.MultiPort(node.PortOut, 1))
	assert.NotNil(t, p)
}

func TestIterateNode_SendAndReceive(t *testing.T) {
	t.Run("Explicit Backward", func(t *testing.T) {
		n := NewIterateNode(1)
		defer n.Close()

		in := port.New()
		inPort := n.Port(node.PortIn)
		inPort.Link(in)

		loop := port.New()
		loopPort := n.Port(node.MultiPort(node.PortOut, 0))
		loopPort.Link(loop)

		done := port.New()
		donePort := n.Port(node.MultiPort(node.PortOut, 1))
		donePort.Link(done)

		proc := process.New()
		defer proc.Exit(nil)

		inStream := in.Open(proc)
		loopStream := loop.Open(proc)
		doneStream := done.Open(proc)

		inPayload := primitive.NewSlice(
			primitive.NewString(faker.UUIDHyphenated()),
			primitive.NewString(faker.UUIDHyphenated()),
		)
		inPck := packet.New(inPayload)

		inStream.Send(inPck)

		ctx, cancel := context.WithTimeout(context.TODO(), time.Second)
		defer cancel()

		for i := 0; i < inPayload.Len(); i++ {
			select {
			case outPck := <-loopStream.Receive():
				assert.Equal(t, inPayload.Get(i), outPck.Payload())
				loopStream.Send(outPck)
			case <-ctx.Done():
				assert.Fail(t, "timeout")
			}
		}

		select {
		case outPck := <-doneStream.Receive():
			assert.Equal(t, inPayload, outPck.Payload())
			doneStream.Send(outPck)
		case <-ctx.Done():
			assert.Fail(t, "timeout")
		}

		select {
		case outPck := <-inStream.Receive():
			assert.Equal(t, inPayload, outPck.Payload())
		case <-ctx.Done():
			assert.Fail(t, "timeout")
		}
	})

	t.Run("Implicit Backward", func(t *testing.T) {
		n := NewIterateNode(1)
		defer n.Close()

		in := port.New()
		inPort := n.Port(node.PortIn)
		inPort.Link(in)

		loop := port.New()
		loopPort := n.Port(node.MultiPort(node.PortOut, 0))
		loopPort.Link(loop)

		proc := process.New()
		defer proc.Exit(nil)

		inStream := in.Open(proc)
		loopStream := loop.Open(proc)

		inPayload := primitive.NewSlice(
			primitive.NewString(faker.UUIDHyphenated()),
			primitive.NewString(faker.UUIDHyphenated()),
		)
		inPck := packet.New(inPayload)

		inStream.Send(inPck)

		ctx, cancel := context.WithTimeout(context.TODO(), time.Second)
		defer cancel()

		for i := 0; i < inPayload.Len(); i++ {
			select {
			case outPck := <-loopStream.Receive():
				assert.Equal(t, inPayload.Get(i), outPck.Payload())
				proc.Stack().Clear(outPck.ID())
			case <-ctx.Done():
				assert.Fail(t, "timeout")
			}
		}

		select {
		case <-proc.Stack().Done(inPck.ID()):
		case <-ctx.Done():
			assert.Fail(t, "timeout")
		}
	})
}
