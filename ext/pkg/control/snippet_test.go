package control

import (
	"context"
	"testing"
	"time"

	"github.com/go-faker/faker/v4"
	"github.com/siyul-park/uniflow/ext/pkg/language"
	"github.com/siyul-park/uniflow/ext/pkg/language/text"
	"github.com/siyul-park/uniflow/pkg/node"
	"github.com/siyul-park/uniflow/pkg/packet"
	"github.com/siyul-park/uniflow/pkg/port"
	"github.com/siyul-park/uniflow/pkg/process"
	"github.com/siyul-park/uniflow/pkg/types"
	"github.com/stretchr/testify/require"
)

func TestSnippetNodeCodec_Compile(t *testing.T) {
	codec := NewSnippetNodeCodec(map[string]language.Compiler{
		text.Language: text.NewCompiler(),
	})

	spec := &SnippetNodeSpec{
		Language: text.Language,
		Code:     "",
	}

	n, err := codec.Compile(spec)
	require.NoError(t, err)
	require.NotNil(t, n)
	require.NoError(t, n.Close())
}

func TestNewSnippetNode(t *testing.T) {
	n := NewSnippetNode(nil)
	require.NotNil(t, n)
	require.NoError(t, n.Close())
}

func TestSnippetNode_SendAndReceive(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.TODO(), time.Second)
	defer cancel()

	n := NewSnippetNode(func(_ context.Context, input any) (any, error) {
		return input, nil
	})
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
	case outPck := <-inWriter.Receive():
		require.Equal(t, inPayload, outPck.Payload())
	case <-ctx.Done():
		require.Fail(t, ctx.Err().Error())
	}
}

func BenchmarkSnippetNode_SendAndReceive(b *testing.B) {
	n := NewSnippetNode(func(_ context.Context, input any) (any, error) {
		return input, nil
	})
	defer n.Close()

	in := port.NewOut()
	in.Link(n.In(node.PortIn))

	proc := process.New()
	defer proc.Exit(nil)

	inWriter := in.Open(proc)

	var inPayload types.Value
	inPck := packet.New(inPayload)

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		inWriter.Write(inPck)
		<-inWriter.Receive()
	}
}
