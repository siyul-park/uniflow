package control

import (
	"context"
	"testing"
	"time"

	"github.com/siyul-park/uniflow/pkg/node"
	"github.com/siyul-park/uniflow/pkg/object"
	"github.com/siyul-park/uniflow/pkg/packet"
	"github.com/siyul-park/uniflow/pkg/port"
	"github.com/siyul-park/uniflow/pkg/process"
	"github.com/stretchr/testify/assert"
)

func TestNewSessionNode(t *testing.T) {
	n := NewSessionNode()
	assert.NotNil(t, n)
	assert.NoError(t, n.Close())
}

func TestSessionNode_Port(t *testing.T) {
	n := NewSessionNode()
	defer n.Close()

	assert.NotNil(t, n.In(node.PortIO))
	assert.NotNil(t, n.In(node.PortIn))
	assert.NotNil(t, n.Out(node.PortOut))
}

func TestSessionNode_SendAndReceive(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.TODO(), time.Second)
	defer cancel()

	n := NewSessionNode()
	defer n.Close()

	io := port.NewOut()
	io.Link(n.In(node.PortIO))

	in := port.NewOut()
	in.Link(n.In(node.PortIn))

	out := port.NewIn()
	n.Out(node.PortOut).Link(out)

	proc := process.New()
	defer proc.Exit(nil)

	ioWriter := io.Open(proc)
	inWriter := in.Open(proc)

	ioPayload := object.NewMap(object.NewString("foo"), object.NewString("bar"))
	ioPck := packet.New(ioPayload)

	ioWriter.Write(ioPck)

	select {
	case outPck := <-ioWriter.Receive():
		assert.Equal(t, packet.None, outPck)
	case <-ctx.Done():
		assert.Fail(t, ctx.Err().Error())
	}

	inPayload := object.NewMap(object.NewString("foo"), object.NewString("baz"))
	inPck := packet.New(inPayload)

	out.AddInitHook(port.InitHookFunc(func(proc *process.Process) {
		outReader := out.Open(proc)

		select {
		case outPck := <-outReader.Read():
			assert.Equal(t, object.NewSlice(ioPayload, inPayload), outPck.Payload())
			outReader.Receive(outPck)
		case <-ctx.Done():
			assert.Fail(t, ctx.Err().Error())
		}
	}))

	inWriter.Write(inPck)

	select {
	case backPck := <-inWriter.Receive():
		assert.NotNil(t, backPck)
	case <-ctx.Done():
		assert.Fail(t, "timeout")
	}
}
