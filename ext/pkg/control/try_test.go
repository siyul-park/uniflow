package control

import (
	"context"
	"testing"
	"time"

	"github.com/go-faker/faker/v4"
	"github.com/pkg/errors"
	"github.com/siyul-park/uniflow/pkg/node"
	"github.com/siyul-park/uniflow/pkg/packet"
	"github.com/siyul-park/uniflow/pkg/port"
	"github.com/siyul-park/uniflow/pkg/process"
	"github.com/siyul-park/uniflow/pkg/types"
	"github.com/stretchr/testify/require"
)

func TestTryNodeCodec_Compile(t *testing.T) {
	codec := NewTryNodeCodec()

	spec := &TryNodeSpec{}

	n, err := codec.Compile(spec)
	require.NoError(t, err)
	require.NotNil(t, n)
	require.NoError(t, n.Close())
}

func TestNewTryNode(t *testing.T) {
	n := NewTryNode()
	require.NotNil(t, n)
	require.NoError(t, n.Close())
}

func TestTryNode_Port(t *testing.T) {
	n := NewTryNode()
	defer n.Close()

	require.NotNil(t, n.In(node.PortIn))
	require.NotNil(t, n.Out(node.PortOut))
	require.NotNil(t, n.Out(node.PortError))
}

func TestTryNode_SendAndReceive(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.TODO(), time.Second)
	defer cancel()

	n1 := NewTryNode()
	defer n1.Close()

	n2 := node.NewOneToOneNode(func(_ *process.Process, inPck *packet.Packet) (*packet.Packet, *packet.Packet) {
		return nil, packet.New(types.NewError(errors.New(faker.Sentence())))
	})
	defer n2.Close()

	n1.Out(node.PortOut).Link(n2.In(node.PortIn))

	in := port.NewOut()
	in.Link(n1.In(node.PortIn))

	err := port.NewIn()
	n1.Out(node.PortError).Link(err)

	proc := process.New()
	defer proc.Exit(nil)

	inWriter := in.Open(proc)
	errReader := err.Open(proc)

	inPayload := types.NewString(faker.UUIDHyphenated())
	inPck := packet.New(inPayload)

	inWriter.Write(inPck)

	select {
	case outPck := <-errReader.Read():
		errReader.Receive(outPck)
	case <-ctx.Done():
		require.Fail(t, ctx.Err().Error())
	}

	select {
	case outPck := <-inWriter.Receive():
		require.IsType(t, outPck.Payload(), types.NewError(nil))
	case <-ctx.Done():
		require.Fail(t, ctx.Err().Error())
	}
}

func BenchmarkTryNode_SendAndReceive(b *testing.B) {
	n1 := NewTryNode()
	defer n1.Close()

	n2 := node.NewOneToOneNode(func(_ *process.Process, inPck *packet.Packet) (*packet.Packet, *packet.Packet) {
		return nil, packet.New(types.NewError(errors.New(faker.Sentence())))
	})
	defer n2.Close()

	n1.Out(node.PortOut).Link(n2.In(node.PortIn))

	in := port.NewOut()
	in.Link(n1.In(node.PortIn))

	proc := process.New()
	defer proc.Exit(nil)

	inWriter := in.Open(proc)

	inPayload := types.NewString(faker.UUIDHyphenated())
	inPck := packet.New(inPayload)

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		inWriter.Write(inPck)
		<-inWriter.Receive()
	}
}
