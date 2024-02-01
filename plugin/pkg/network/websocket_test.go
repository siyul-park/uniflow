package network

import (
	"context"
	"testing"
	"time"

	"github.com/siyul-park/uniflow/pkg/node"
	"github.com/siyul-park/uniflow/pkg/packet"
	"github.com/siyul-park/uniflow/pkg/port"
	"github.com/siyul-park/uniflow/pkg/primitive"
	"github.com/siyul-park/uniflow/pkg/process"
	"github.com/stretchr/testify/assert"
)

func TestNewWebsocketNode(t *testing.T) {
	n := NewWebsocketNode()
	assert.NotNil(t, n)
	assert.NoError(t, n.Close())
}

func TestWebsocketNode_Port(t *testing.T) {
	n := NewWebsocketNode()
	defer n.Close()

	p, ok := n.Port(node.PortIO)
	assert.True(t, ok)
	assert.NotNil(t, p)

	p, ok = n.Port(node.PortIn)
	assert.True(t, ok)
	assert.NotNil(t, p)

	p, ok = n.Port(node.PortOut)
	assert.True(t, ok)
	assert.NotNil(t, p)

	p, ok = n.Port(node.PortErr)
	assert.True(t, ok)
	assert.NotNil(t, p)
}

func TestWebsocketNode_Upgrade(t *testing.T) {
	n := NewWebsocketNode()
	defer n.Close()

	io := port.New()
	ioPort, _ := n.Port(node.PortIO)
	ioPort.Link(io)

	proc := process.New()
	defer proc.Exit(nil)

	ioStream := io.Open(proc)

	var inPayload primitive.Value
	inPck := packet.New(inPayload)

	ioStream.Send(inPck)

	ctx, cancel := context.WithTimeout(context.TODO(), time.Second)
	defer cancel()

	select {
	case outPck := <-ioStream.Receive():
		assert.Equal(t, primitive.NewMap(), outPck.Payload())
	case <-ctx.Done():
		assert.Fail(t, ctx.Err().Error())
	}
}
