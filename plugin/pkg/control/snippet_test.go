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
	"github.com/stretchr/testify/assert"
)

func TestSnippetNodeCodec_Decode(t *testing.T) {
	codec := NewSnippetNodeCodec()

	spec := &SnippetNodeSpec{
		Lang: LangJSON,
		Code: "{}",
	}

	n, err := codec.Decode(spec)
	assert.NoError(t, err)
	assert.NotNil(t, n)
}

func TestNewSnippetNode(t *testing.T) {
	t.Run(LangJSON, func(t *testing.T) {
		n, err := NewSnippetNode(LangJSON, `{}`)
		assert.NoError(t, err)
		assert.NotNil(t, n)

		assert.NoError(t, n.Close())
	})

	t.Run(LangYAML, func(t *testing.T) {
		n, err := NewSnippetNode(LangYAML, `{}`)
		assert.NoError(t, err)
		assert.NotNil(t, n)

		assert.NoError(t, n.Close())
	})

	t.Run(LangJSONata, func(t *testing.T) {
		n, err := NewSnippetNode(LangJSONata, `$`)
		assert.NoError(t, err)
		assert.NotNil(t, n)

		assert.NoError(t, n.Close())
	})
}

func TestSnippetNode_SendAndReceive(t *testing.T) {
	t.Run(LangJSON, func(t *testing.T) {
		n, _ := NewSnippetNode(LangJSON, `{}`)
		defer n.Close()

		io := port.New()
		ioPort, _ := n.Port(node.PortIO)
		ioPort.Link(io)

		proc := process.New()
		defer proc.Exit(nil)

		ioStream := io.Open(proc)

		var inPayload primitive.Value
		inPck := packet.New(inPayload)

		ioStream.Send(inPck)

		ctx, cancel := context.WithTimeout(context.TODO(), time.Second)
		defer cancel()

		select {
		case outPck := <-ioStream.Receive():
			assert.Equal(t, primitive.NewMap(), outPck.Payload())
		case <-ctx.Done():
			assert.Fail(t, ctx.Err().Error())
		}
	})

	t.Run(LangYAML, func(t *testing.T) {
		n, _ := NewSnippetNode(LangYAML, `{}`)
		defer n.Close()

		io := port.New()
		ioPort, _ := n.Port(node.PortIO)
		ioPort.Link(io)

		proc := process.New()
		defer proc.Exit(nil)

		ioStream := io.Open(proc)

		var inPayload primitive.Value
		inPck := packet.New(inPayload)

		ioStream.Send(inPck)

		ctx, cancel := context.WithTimeout(context.TODO(), time.Second)
		defer cancel()

		select {
		case outPck := <-ioStream.Receive():
			assert.Equal(t, primitive.NewMap(), outPck.Payload())
		case <-ctx.Done():
			assert.Fail(t, ctx.Err().Error())
		}
	})

	t.Run(LangJSONata, func(t *testing.T) {
		n, _ := NewSnippetNode(LangJSONata, `$`)
		defer n.Close()

		io := port.New()
		ioPort, _ := n.Port(node.PortIO)
		ioPort.Link(io)

		proc := process.New()
		defer proc.Exit(nil)

		ioStream := io.Open(proc)

		inPayload := primitive.NewString(faker.Word())
		inPck := packet.New(inPayload)

		ioStream.Send(inPck)

		ctx, cancel := context.WithTimeout(context.TODO(), time.Second)
		defer cancel()

		select {
		case outPck := <-ioStream.Receive():
			assert.Equal(t, inPayload, outPck.Payload())
		case <-ctx.Done():
			assert.Fail(t, ctx.Err().Error())
		}
	})
}

func BenchmarkSnippetNode_SendAndReceive(b *testing.B) {
	b.Run(LangJSON, func(b *testing.B) {
		n, _ := NewSnippetNode(LangJSON, "{}")
		defer n.Close()

		io := port.New()
		ioPort, _ := n.Port(node.PortIO)
		ioPort.Link(io)

		proc := process.New()
		defer proc.Exit(nil)

		ioStream := io.Open(proc)

		var inPayload primitive.Value
		inPck := packet.New(inPayload)

		ioStream.Send(inPck)

		b.ResetTimer()

		b.RunParallel(func(p *testing.PB) {
			for p.Next() {
				ioStream.Send(inPck)
				<-ioStream.Receive()
			}
		})
	})

	b.Run(LangYAML, func(b *testing.B) {
		n, _ := NewSnippetNode(LangYAML, "{}")
		defer n.Close()

		io := port.New()
		ioPort, _ := n.Port(node.PortIO)
		ioPort.Link(io)

		proc := process.New()
		defer proc.Exit(nil)

		ioStream := io.Open(proc)

		var inPayload primitive.Value
		inPck := packet.New(inPayload)

		ioStream.Send(inPck)

		b.ResetTimer()

		b.RunParallel(func(p *testing.PB) {
			for p.Next() {
				ioStream.Send(inPck)
				<-ioStream.Receive()
			}
		})
	})

	b.Run(LangJSONata, func(b *testing.B) {
		n, _ := NewSnippetNode(LangJSONata, "$")
		defer n.Close()

		io := port.New()
		ioPort, _ := n.Port(node.PortIO)
		ioPort.Link(io)

		proc := process.New()
		defer proc.Exit(nil)

		ioStream := io.Open(proc)

		inPayload := primitive.NewString(faker.Word())
		inPck := packet.New(inPayload)

		ioStream.Send(inPck)

		b.ResetTimer()

		b.RunParallel(func(p *testing.PB) {
			for p.Next() {
				ioStream.Send(inPck)
				<-ioStream.Receive()
			}
		})
	})
}
