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

func TestNewForeachNode(t *testing.T) {
	n := NewForeachNode(0)
	assert.NotNil(t, n)

	assert.NoError(t, n.Close())
}

func TestForeachNode_Port(t *testing.T) {
	n := NewForeachNode(0)
	defer n.Close()

	p := n.Port(node.PortIO)
	assert.NotNil(t, p)

	p = n.Port(node.PortIn)
	assert.NotNil(t, p)

	p = n.Port(node.PortOut)
	assert.NotNil(t, p)
}

func TestForeachNode_SendAndReceive(t *testing.T) {
	n := NewForeachNode(1)
	defer n.Close()

	io := port.New()
	ioPort := n.Port(node.PortIO)
	ioPort.Link(io)

	in := port.New()
	inPort := n.Port(node.PortIn)
	inPort.Link(in)

	out := port.New()
	outPort := n.Port(node.PortOut)
	outPort.Link(out)

	proc := process.New()
	defer proc.Exit(nil)

	ioStream := io.Open(proc)
	inStream := in.Open(proc)
	outStream := out.Open(proc)

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
		case outPck := <-ioStream.Receive():
			assert.Equal(t, inPayload.Get(i), outPck.Payload())
			ioStream.Send(outPck)
		case <-ctx.Done():
			assert.Fail(t, "timeout")
		}
	}

	select {
	case outPck := <-outStream.Receive():
		assert.Equal(t, inPayload, outPck.Payload())
		outStream.Send(outPck)
	case <-ctx.Done():
		assert.Fail(t, "timeout")
	}

	select {
	case outPck := <-inStream.Receive():
		assert.Equal(t, inPayload, outPck.Payload())
	case <-ctx.Done():
		assert.Fail(t, "timeout")
	}
}
