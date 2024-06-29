package control

import (
	"context"
	"testing"
	"time"

	"github.com/go-faker/faker/v4"
	"github.com/siyul-park/uniflow/extend/pkg/language"
	"github.com/siyul-park/uniflow/extend/pkg/language/text"
	"github.com/siyul-park/uniflow/pkg/node"
	"github.com/siyul-park/uniflow/pkg/object"
	"github.com/siyul-park/uniflow/pkg/packet"
	"github.com/siyul-park/uniflow/pkg/port"
	"github.com/siyul-park/uniflow/pkg/process"
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

	inPayload := object.NewString(faker.Word())
	inPck := packet.New(inPayload)

	inWriter.Write(inPck)

	select {
	case outPck := <-inWriter.Receive():
		assert.Equal(t, inPayload, outPck.Payload())
	case <-ctx.Done():
		assert.Fail(t, ctx.Err().Error())
	}
}

func TestSnippetNodeCodec_Decode(t *testing.T) {
	m := language.NewModule()
	m.Store(text.Language, text.NewCompiler())

	codec := NewSnippetNodeCodec(m)

	spec := &SnippetNodeSpec{
		Language: text.Language,
		Code:     "",
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

	var inPayload object.Object
	inPck := packet.New(inPayload)

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		inWriter.Write(inPck)
		<-inWriter.Receive()
	}
}
