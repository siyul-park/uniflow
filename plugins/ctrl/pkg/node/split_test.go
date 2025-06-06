package node

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

func TestSplitNodeCodec_Compile(t *testing.T) {
	codec := NewSplitNodeCodec()

	spec := &SplitNodeSpec{}

	n, err := codec.Compile(spec)
	require.NoError(t, err)
	require.NotNil(t, n)
	require.NoError(t, n.Close())
}

func TestNewSplitNode(t *testing.T) {
	n := NewSplitNode()
	require.NotNil(t, n)
	require.NoError(t, n.Close())
}

func TestSplitNode_SendAndReceive(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.TODO(), time.Second)
	defer cancel()

	n := NewSplitNode()
	defer n.Close()

	in := port.NewOut()
	in.Link(n.In(node.PortIn))

	var outs []*port.InPort
	for i := 0; i < 4; i++ {
		out := port.NewIn()
		n.Out(node.PortWithIndex(node.PortOut, i)).Link(out)
		outs = append(outs, out)
	}

	proc := process.New()
	defer proc.Exit(nil)

	inWriter := in.Open(proc)
	var outReaders []*packet.Reader
	for _, out := range outs {
		outReader := out.Open(proc)
		outReaders = append(outReaders, outReader)
	}

	inPayload := types.NewSlice()
	for range outs {
		inPayload = inPayload.Append(types.NewString(faker.UUIDHyphenated()))
	}
	inPck := packet.New(inPayload)

	inWriter.Write(inPck)

	for _, outReader := range outReaders {
		select {
		case outPck := <-outReader.Read():
			outReader.Receive(outPck)
		case <-ctx.Done():
			require.Fail(t, ctx.Err().Error())
		}
	}

	select {
	case backPck := <-inWriter.Receive():
		require.NotNil(t, backPck)
	case <-ctx.Done():
		require.Fail(t, ctx.Err().Error())
	}
}

func BenchmarkSplitNode_SendAndReceive(b *testing.B) {
	n := NewSplitNode()
	defer n.Close()

	in := port.NewOut()
	in.Link(n.In(node.PortIn))

	var outs []*port.InPort
	for i := 0; i < 4; i++ {
		out := port.NewIn()
		n.Out(node.PortWithIndex(node.PortOut, i)).Link(out)
		outs = append(outs, out)
	}

	proc := process.New()
	defer proc.Exit(nil)

	inWriter := in.Open(proc)
	var outReaders []*packet.Reader
	for _, out := range outs {
		outReader := out.Open(proc)
		outReaders = append(outReaders, outReader)
	}

	inPayload := types.NewSlice()
	for range outs {
		inPayload = inPayload.Append(types.NewString(faker.UUIDHyphenated()))
	}
	inPck := packet.New(inPayload)

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		inWriter.Write(inPck)

		for _, outReader := range outReaders {
			outPck := <-outReader.Read()
			outReader.Receive(outPck)
		}
	}
}
