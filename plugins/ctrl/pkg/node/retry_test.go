package node

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

func TestRetryNodeCodec_Compile(t *testing.T) {
	codec := NewRetryNodeCodec()

	spec := &RetryNodeSpec{
		Threshold: 0,
	}

	n, err := codec.Compile(spec)
	require.NoError(t, err)
	require.NotNil(t, n)
	require.NoError(t, n.Close())
}

func TestNewRetryNode(t *testing.T) {
	n := NewRetryNode(0)
	require.NotNil(t, n)
	require.NoError(t, n.Close())
}

func TestRetryNode_Port(t *testing.T) {
	n := NewRetryNode(0)
	defer n.Close()

	require.NotNil(t, n.In(node.PortIn))
	require.NotNil(t, n.Out(node.PortOut))
}

func TestRetryNode_SendAndReceive(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.TODO(), time.Second)
	defer cancel()

	limit := 2

	n1 := NewRetryNode(limit)
	defer n1.Close()

	count := 0
	n2 := node.NewOneToOneNode(func(_ *process.Process, inPck *packet.Packet) (*packet.Packet, *packet.Packet) {
		count += 1
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

	inWriter.Write(inPck)

	select {
	case outPck := <-inWriter.Receive():
		require.Equal(t, limit+1, count)
		require.IsType(t, outPck.Payload(), types.NewError(nil))
	case <-ctx.Done():
		require.Fail(t, ctx.Err().Error())
	}
}

func BenchmarkRetryNode_SendAndReceive(b *testing.B) {
	n1 := NewRetryNode(1)
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
