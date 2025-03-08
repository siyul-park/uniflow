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

func TestMergeNodeCodec_Compile(t *testing.T) {
	codec := NewMergeNodeCodec()

	spec := &MergeNodeSpec{}

	n, err := codec.Compile(spec)
	require.NoError(t, err)
	require.NotNil(t, n)
	require.NoError(t, n.Close())
}

func TestNewMergeNode(t *testing.T) {
	n := NewMergeNode()
	require.NotNil(t, n)
	require.NoError(t, n.Close())
}

func TestMergeNode_SendAndReceive(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.TODO(), time.Second)
	defer cancel()

	n := NewMergeNode()
	defer n.Close()

	var ins []*port.OutPort
	for i := 0; i < 4; i++ {
		in := port.NewOut()
		in.Link(n.In(node.PortWithIndex(node.PortIn, i)))
		ins = append(ins, in)
	}

	out := port.NewIn()
	n.Out(node.PortOut).Link(out)

	proc := process.New()
	defer proc.Exit(nil)

	inWriters := make([]*packet.Writer, len(ins))
	for i, in := range ins {
		inWriters[i] = in.Open(proc)
	}
	outReader := out.Open(proc)

	var inPayloads []types.Value
	for range inWriters {
		inPayloads = append(inPayloads, types.NewString(faker.UUIDHyphenated()))
	}

	for i, inWriter := range inWriters {
		inPck := packet.New(inPayloads[i])
		inWriter.Write(inPck)
	}

	select {
	case outPck := <-outReader.Read():
		require.Equal(t, types.NewSlice(inPayloads...), outPck.Payload())
		outReader.Receive(outPck)
	case <-ctx.Done():
		require.Fail(t, ctx.Err().Error())
	}

	for _, inWriter := range inWriters {
		select {
		case backPck := <-inWriter.Receive():
			require.NotNil(t, backPck)
		case <-ctx.Done():
			require.Fail(t, ctx.Err().Error())
		}
	}
}

func BenchmarkMergeNode_SendAndReceive(b *testing.B) {
	n := NewMergeNode()
	defer n.Close()

	var ins []*port.OutPort
	for i := 0; i < 4; i++ {
		in := port.NewOut()
		in.Link(n.In(node.PortWithIndex(node.PortIn, i)))
		ins = append(ins, in)
	}

	out := port.NewIn()
	n.Out(node.PortOut).Link(out)

	proc := process.New()
	defer proc.Exit(nil)

	inWriters := make([]*packet.Writer, len(ins))
	for i, in := range ins {
		inWriters[i] = in.Open(proc)
	}
	outReader := out.Open(proc)

	var inPayloads []types.Value
	for range inWriters {
		inPayloads = append(inPayloads, types.NewString(faker.UUIDHyphenated()))
	}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		for i, inWriter := range inWriters {
			inPck := packet.New(inPayloads[i])
			inWriter.Write(inPck)
		}

		outPck := <-outReader.Read()
		outReader.Receive(outPck)

		for _, inWriter := range inWriters {
			<-inWriter.Receive()
		}
	}
}
