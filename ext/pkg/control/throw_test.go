package control

import (
	"context"
	"testing"
	"time"

	"github.com/go-faker/faker/v4"
	"github.com/siyul-park/uniflow/pkg/node"
	"github.com/siyul-park/uniflow/pkg/packet"
	"github.com/siyul-park/uniflow/pkg/port"
	"github.com/siyul-park/uniflow/pkg/process"
	"github.com/siyul-park/uniflow/pkg/types"
	"github.com/stretchr/testify/require"
)

func TestThrowNodeCodec_Compile(t *testing.T) {
	codec := NewThrowNodeCodec()

	spec := &ThrowNodeSpec{}

	n, err := codec.Compile(spec)
	require.NoError(t, err)
	require.NotNil(t, n)
	require.NoError(t, n.Close())
}

func TestNewThrowNode(t *testing.T) {
	n := NewThrowNode()
	require.NotNil(t, n)
	require.NoError(t, n.Close())
}

func TestThrowNode_Port(t *testing.T) {
	n := NewThrowNode()
	defer n.Close()

	require.NotNil(t, n.In(node.PortIn))
}

func TestThrowNode_SendAndReceive(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.TODO(), time.Second)
	defer cancel()

	n := NewThrowNode()
	defer n.Close()

	in := port.NewOut()
	in.Link(n.In(node.PortIn))

	proc := process.New()
	defer proc.Exit(nil)

	inWriter := in.Open(proc)

	inPayload := types.NewString(faker.UUIDHyphenated())
	inPck := packet.New(inPayload)

	inWriter.Write(inPck)

	select {
	case backPck := <-inWriter.Receive():
		require.IsType(t, backPck.Payload(), types.NewError(nil))
	case <-ctx.Done():
		require.Fail(t, ctx.Err().Error())
	}
}

func BenchmarkThrowNode_SendAndReceive(b *testing.B) {
	n := NewThrowNode()
	defer n.Close()

	in := port.NewOut()
	in.Link(n.In(node.PortIn))

	proc := process.New()
	defer proc.Exit(nil)

	inWriter := in.Open(proc)

	inPayload := types.NewString(faker.UUIDHyphenated())
	inPck := packet.New(inPayload)

	inWriter.Write(inPck)

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		inWriter.Write(inPck)
		<-inWriter.Receive()
	}
}
