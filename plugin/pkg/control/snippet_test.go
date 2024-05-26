package control

import (
	"context"
	"testing"
	"time"

	"github.com/go-faker/faker/v4"
	"github.com/siyul-park/uniflow/pkg/node"
	"github.com/siyul-park/uniflow/pkg/packet"
	"github.com/siyul-park/uniflow/pkg/port"
	"github.com/siyul-park/uniflow/pkg/primitive"
	"github.com/siyul-park/uniflow/pkg/process"
	"github.com/siyul-park/uniflow/plugin/internal/language"
	"github.com/stretchr/testify/assert"
)

func TestNewSnippetNode(t *testing.T) {
	n := NewSnippetNode(nil)
	assert.NotNil(t, n)
	assert.NoError(t, n.Close())
}

func TestSnippetNode_SendAndReceive(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.TODO(), time.Second)
	defer cancel()

	n := NewSnippetNode(func(input any) (any, error) {
		return input, nil
	})
	defer n.Close()

	in := port.NewOut()
	in.Link(n.In(node.PortIn))

	proc := process.New()
	defer proc.Exit(nil)

	inWriter := in.Open(proc)

	inPayload := primitive.NewString(faker.Word())
	inPck := packet.New(inPayload)

	inWriter.Write(inPck)

	select {
	case outPck := <-inWriter.Receive():
		assert.Equal(t, inPayload.Interface(), outPck.Payload().Interface())
	case <-ctx.Done():
		assert.Fail(t, ctx.Err().Error())
	}
}

func TestSnippetNodeCodec_Decode(t *testing.T) {
	codec := NewSnippetNodeCodec()

	spec := &SnippetNodeSpec{
		Lang: language.Text,
		Code: "",
	}

	n, err := codec.Decode(spec)
	assert.NoError(t, err)
	assert.NotNil(t, n)
	assert.NoError(t, n.Close())
}

func BenchmarkSnippetNode_SendAndReceive(b *testing.B) {
	n := NewSnippetNode(func(input any) (any, error) {
		return input, nil
	})
	defer n.Close()

	in := port.NewOut()
	in.Link(n.In(node.PortIn))

	proc := process.New()
	defer proc.Exit(nil)

	inWriter := in.Open(proc)

	var inPayload primitive.Value
	inPck := packet.New(inPayload)

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		inWriter.Write(inPck)
		<-inWriter.Receive()
	}
}
