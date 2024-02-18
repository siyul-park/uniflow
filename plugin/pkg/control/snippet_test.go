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

func TestNewSnippetNode(t *testing.T) {
	t.Run(language.Text, func(t *testing.T) {
		n, err := NewSnippetNode(language.Text, "")
		assert.NoError(t, err)
		assert.NotNil(t, n)

		assert.NoError(t, n.Close())
	})

	t.Run(language.Typescript, func(t *testing.T) {
		n, err := NewSnippetNode(language.Typescript, `export default function (input: any): any {
			return input;
		}`)
		assert.NoError(t, err)
		assert.NotNil(t, n)

		assert.NoError(t, n.Close())
	})

	t.Run(language.JSON, func(t *testing.T) {
		n, err := NewSnippetNode(language.JSON, `{}`)
		assert.NoError(t, err)
		assert.NotNil(t, n)

		assert.NoError(t, n.Close())
	})

	t.Run(language.YAML, func(t *testing.T) {
		n, err := NewSnippetNode(language.YAML, `{}`)
		assert.NoError(t, err)
		assert.NotNil(t, n)

		assert.NoError(t, n.Close())
	})

	t.Run(language.Javascript, func(t *testing.T) {
		n, err := NewSnippetNode(language.Javascript, `export default function (input) {
			return input;
		}`)
		assert.NoError(t, err)
		assert.NotNil(t, n)

		assert.NoError(t, n.Close())
	})

	t.Run(language.JSONata, func(t *testing.T) {
		n, err := NewSnippetNode(language.JSONata, `$`)
		assert.NoError(t, err)
		assert.NotNil(t, n)

		assert.NoError(t, n.Close())
	})
}

func TestSnippetNode_SendAndReceive(t *testing.T) {
	t.Run(language.Text, func(t *testing.T) {
		n, _ := NewSnippetNode(language.Text, "")
		defer n.Close()

		io := port.New()
		ioPort := n.Port(node.PortIO)
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
			assert.Equal(t, primitive.NewString(""), outPck.Payload())
		case <-ctx.Done():
			assert.Fail(t, ctx.Err().Error())
		}
	})

	t.Run(language.Typescript, func(t *testing.T) {
		n, _ := NewSnippetNode(language.Typescript, `export default function (input: any): any {
			return input;
		}`)
		defer n.Close()

		io := port.New()
		ioPort := n.Port(node.PortIO)
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

	t.Run(language.JSON, func(t *testing.T) {
		n, _ := NewSnippetNode(language.JSON, `{}`)
		defer n.Close()

		io := port.New()
		ioPort := n.Port(node.PortIO)
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

	t.Run(language.YAML, func(t *testing.T) {
		n, _ := NewSnippetNode(language.YAML, `{}`)
		defer n.Close()

		io := port.New()
		ioPort := n.Port(node.PortIO)
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

	t.Run(language.Javascript, func(t *testing.T) {
		n, _ := NewSnippetNode(language.Javascript, `export default function (input) {
			return input;
		}`)
		defer n.Close()

		io := port.New()
		ioPort := n.Port(node.PortIO)
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

	t.Run(language.JSONata, func(t *testing.T) {
		n, _ := NewSnippetNode(language.JSONata, `$`)
		defer n.Close()

		io := port.New()
		ioPort := n.Port(node.PortIO)
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
	b.Run(language.Text, func(b *testing.B) {
		n, _ := NewSnippetNode(language.Text, "")
		defer n.Close()

		io := port.New()
		ioPort := n.Port(node.PortIO)
		ioPort.Link(io)

		proc := process.New()
		defer proc.Exit(nil)

		ioStream := io.Open(proc)

		var inPayload primitive.Value
		inPck := packet.New(inPayload)

		b.ResetTimer()

		b.RunParallel(func(p *testing.PB) {
			for p.Next() {
				ioStream.Send(inPck)
				<-ioStream.Receive()
			}
		})
	})

	b.Run(language.Typescript, func(b *testing.B) {
		n, _ := NewSnippetNode(language.Typescript, `export default function (input: any): any {
			return input;
		}`)

		io := port.New()
		ioPort := n.Port(node.PortIO)
		ioPort.Link(io)

		proc := process.New()
		defer proc.Exit(nil)

		ioStream := io.Open(proc)

		inPayload := primitive.NewString(faker.Word())
		inPck := packet.New(inPayload)

		b.ResetTimer()

		b.RunParallel(func(p *testing.PB) {
			for p.Next() {
				ioStream.Send(inPck)
				<-ioStream.Receive()
			}
		})
	})

	b.Run(language.JSON, func(b *testing.B) {
		n, _ := NewSnippetNode(language.JSON, "{}")
		defer n.Close()

		io := port.New()
		ioPort := n.Port(node.PortIO)
		ioPort.Link(io)

		proc := process.New()
		defer proc.Exit(nil)

		ioStream := io.Open(proc)

		var inPayload primitive.Value
		inPck := packet.New(inPayload)

		b.ResetTimer()

		b.RunParallel(func(p *testing.PB) {
			for p.Next() {
				ioStream.Send(inPck)
				<-ioStream.Receive()
			}
		})
	})

	b.Run(language.YAML, func(b *testing.B) {
		n, _ := NewSnippetNode(language.YAML, "{}")
		defer n.Close()

		io := port.New()
		ioPort := n.Port(node.PortIO)
		ioPort.Link(io)

		proc := process.New()
		defer proc.Exit(nil)

		ioStream := io.Open(proc)

		var inPayload primitive.Value
		inPck := packet.New(inPayload)

		b.ResetTimer()

		b.RunParallel(func(p *testing.PB) {
			for p.Next() {
				ioStream.Send(inPck)
				<-ioStream.Receive()
			}
		})
	})

	b.Run(language.Javascript, func(b *testing.B) {
		n, _ := NewSnippetNode(language.Javascript, `export default function (input) {
			return input;
		}`)

		io := port.New()
		ioPort := n.Port(node.PortIO)
		ioPort.Link(io)

		proc := process.New()
		defer proc.Exit(nil)

		ioStream := io.Open(proc)

		inPayload := primitive.NewString(faker.Word())
		inPck := packet.New(inPayload)

		b.ResetTimer()

		b.RunParallel(func(p *testing.PB) {
			for p.Next() {
				ioStream.Send(inPck)
				<-ioStream.Receive()
			}
		})
	})

	b.Run(language.JSONata, func(b *testing.B) {
		n, _ := NewSnippetNode(language.JSONata, "$")
		defer n.Close()

		io := port.New()
		ioPort := n.Port(node.PortIO)
		ioPort.Link(io)

		proc := process.New()
		defer proc.Exit(nil)

		ioStream := io.Open(proc)

		inPayload := primitive.NewString(faker.Word())
		inPck := packet.New(inPayload)

		b.ResetTimer()

		b.RunParallel(func(p *testing.PB) {
			for p.Next() {
				ioStream.Send(inPck)
				<-ioStream.Receive()
			}
		})
	})
}
