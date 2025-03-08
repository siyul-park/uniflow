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
	"github.com/stretchr/testify/require"
)

func TestForkNodeCodec_Compile(t *testing.T) {
	codec := NewForkNodeCodec()

	spec := &ForkNodeSpec{}

	n, err := codec.Compile(spec)
	require.NoError(t, err)
	require.NotNil(t, n)
	require.NoError(t, n.Close())
}

func TestNewForkNode(t *testing.T) {
	n := NewForkNode()
	require.NotNil(t, n)
	require.NoError(t, n.Close())
}

func TestForkNode_Port(t *testing.T) {
	n := NewForkNode()
	defer n.Close()

	require.NotNil(t, n.In(node.PortIn))
	require.NotNil(t, n.Out(node.PortOut))
	require.NotNil(t, n.Out(node.PortError))
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
	defer proc.Join()

	inWriter := in.Open(proc)

	inPayload := types.NewMap(types.NewString("foo"), types.NewString("bar"))
	inPck := packet.New(inPayload)

	out.AddListener(port.ListenFunc(func(proc *process.Process) {
		outReader := out.Open(proc)

		select {
		case outPck := <-outReader.Read():
			require.Equal(t, inPayload, outPck.Payload())
			outReader.Receive(outPck)
		case <-ctx.Done():
			require.Fail(t, ctx.Err().Error())
		}
	}))

	inWriter.Write(inPck)

	select {
	case backPck := <-inWriter.Receive():
		require.NotNil(t, backPck)
	case <-ctx.Done():
		require.Fail(t, ctx.Err().Error())
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
	defer proc.Join()

	inWriter := in.Open(proc)

	inPayload := types.NewMap(types.NewString("foo"), types.NewString("bar"))
	inPck := packet.New(inPayload)

	out.AddListener(port.ListenFunc(func(proc *process.Process) {
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
