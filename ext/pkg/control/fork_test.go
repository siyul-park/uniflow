package control

import (
	"context"
	"testing"
	"time"

	"github.com/siyul-park/uniflow/pkg/node"
	"github.com/siyul-park/uniflow/pkg/packet"
	"github.com/siyul-park/uniflow/pkg/port"
	"github.com/siyul-park/uniflow/pkg/process"
	"github.com/siyul-park/uniflow/pkg/types"
	"github.com/stretchr/testify/assert"
)

func TestForkNodeCodec_Decode(t *testing.T) {
	codec := NewForkNodeCodec()

	spec := &ForkNodeSpec{}

	n, err := codec.Compile(spec)
	assert.NoError(t, err)
	assert.NotNil(t, n)
	assert.NoError(t, n.Close())
}

func TestNewForkNode(t *testing.T) {
	n := NewForkNode()
	assert.NotNil(t, n)
	assert.NoError(t, n.Close())
}

func TestForkNode_Port(t *testing.T) {
	n := NewForkNode()
	defer n.Close()

	assert.NotNil(t, n.In(node.PortIn))
	assert.NotNil(t, n.Out(node.PortOut))
	assert.NotNil(t, n.Out(node.PortErr))
}

func TestForkNode_SendAndReceive(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.TODO(), time.Second)
	defer cancel()

	n := NewForkNode()
	defer n.Close()

	in := port.NewOut()
	in.Link(n.In(node.PortIn))

	out := port.NewIn()
	n.Out(node.PortOut).Link(out)

	proc := process.New()
	defer proc.Exit(nil)
	defer proc.Wait()

	inWriter := in.Open(proc)

	inPayload := types.NewMap(types.NewString("foo"), types.NewString("bar"))
	inPck := packet.New(inPayload)

	out.Accept(port.ListenFunc(func(proc *process.Process) {
		outReader := out.Open(proc)

		select {
		case outPck := <-outReader.Read():
			assert.Equal(t, inPayload, outPck.Payload())
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

func BenchmarkForkNode_SendAndReceive(b *testing.B) {
	n := NewForkNode()
	defer n.Close()

	in := port.NewOut()
	in.Link(n.In(node.PortIn))

	out := port.NewIn()
	n.Out(node.PortOut).Link(out)

	proc := process.New()
	defer proc.Exit(nil)
	defer proc.Wait()

	inWriter := in.Open(proc)

	inPayload := types.NewMap(types.NewString("foo"), types.NewString("bar"))
	inPck := packet.New(inPayload)

	out.Accept(port.ListenFunc(func(proc *process.Process) {
		outReader := out.Open(proc)

		for {
			outPck, ok := <-outReader.Read()
			if !ok {
				return
			}

			outReader.Receive(outPck)
		}
	}))

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		inWriter.Write(inPck)
		<-inWriter.Receive()
	}
}
