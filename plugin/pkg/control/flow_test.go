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

func TestNewFlowNode(t *testing.T) {
	n := NewFlowNode()
	assert.NotNil(t, n)

	assert.NoError(t, n.Close())
}

func TestFlowNode_Port(t *testing.T) {
	n := NewFlowNode()
	defer n.Close()

	p, ok := n.Port(node.PortIn)
	assert.True(t, ok)
	assert.NotNil(t, p)

	p, ok = n.Port(node.PortOut)
	assert.True(t, ok)
	assert.NotNil(t, p)
}

func TestFlowNode_SendAndReceive(t *testing.T) {
	n := NewFlowNode()
	defer n.Close()

	in := port.New()
	inPort, _ := n.Port(node.PortIn)
	inPort.Link(in)

	out := port.New()
	outPort, _ := n.Port(node.PortOut)
	outPort.Link(out)

	proc := process.New()
	defer proc.Exit(nil)

	inStream := in.Open(proc)
	outStream := out.Open(proc)

	inPayload := primitive.NewSlice(primitive.NewString(faker.UUIDHyphenated()), primitive.NewString(faker.UUIDHyphenated()))
	inPck := packet.New(inPayload)

	inStream.Send(inPck)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	for i := 0; i < inPayload.Len(); i++ {
		select {
		case outPck := <-outStream.Receive():
			assert.Equal(t, inPayload.Get(i), outPck.Payload())
		case <-ctx.Done():
			assert.Fail(t, "timeout")
		}
	}
}
