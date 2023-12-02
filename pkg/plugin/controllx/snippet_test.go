package controllx

import (
	"context"
	"fmt"
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

func TestNewSnippetNode(t *testing.T) {
	n, err := NewSnippetNode(SnippetNodeConfig{
		Lang: LangJSON,
		Code: "{}",
	})
	assert.NoError(t, err)
	assert.NotNil(t, n)
	assert.NotZero(t, n.ID())

	_ = n.Close()
}

func TestSnippetNode_Send(t *testing.T) {
	t.Run(LangTypescript, func(t *testing.T) {
		n, _ := NewSnippetNode(SnippetNodeConfig{
			Lang: LangTypescript,
			Code: `
function main(inPayload: any): any {
	return inPayload;
}
			`,
		})
		defer func() { _ = n.Close() }()

		io := port.New()
		ioPort, _ := n.Port(node.PortIO)
		ioPort.Link(io)

		proc := process.New()
		defer proc.Exit(nil)

		ioStream := io.Open(proc)

		inPayload := primitive.NewString(faker.UUIDHyphenated())
		inPck := packet.New(inPayload)

		ioStream.Send(inPck)

		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()

		select {
		case outPck := <-ioStream.Receive():
			assert.Equal(t, inPayload, outPck.Payload())
		case <-ctx.Done():
			assert.Fail(t, "timeout")
		}
	})

	t.Run(LangJavascript, func(t *testing.T) {
		n, _ := NewSnippetNode(SnippetNodeConfig{
			Lang: LangJavascript,
			Code: `
function main(inPayload) {
	return inPayload;
}
			`,
		})
		defer func() { _ = n.Close() }()

		io := port.New()
		ioPort, _ := n.Port(node.PortIO)
		ioPort.Link(io)

		proc := process.New()
		defer proc.Exit(nil)

		ioStream := io.Open(proc)

		inPayload := primitive.NewString(faker.UUIDHyphenated())
		inPck := packet.New(inPayload)

		ioStream.Send(inPck)

		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()

		select {
		case outPck := <-ioStream.Receive():
			assert.Equal(t, inPayload, outPck.Payload())
		case <-ctx.Done():
			assert.Fail(t, "timeout")
		}
	})

	t.Run(LangJSON, func(t *testing.T) {
		data := faker.UUIDHyphenated()

		n, _ := NewSnippetNode(SnippetNodeConfig{
			Lang: LangJSON,
			Code: fmt.Sprintf("\"%s\"", data),
		})
		defer func() { _ = n.Close() }()

		io := port.New()
		ioPort, _ := n.Port(node.PortIO)
		ioPort.Link(io)

		proc := process.New()
		defer proc.Exit(nil)

		ioStream := io.Open(proc)

		inPayload := primitive.NewString(faker.UUIDHyphenated())
		inPck := packet.New(inPayload)

		ioStream.Send(inPck)

		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()

		select {
		case outPck := <-ioStream.Receive():
			assert.Equal(t, data, outPck.Payload().Interface())
		case <-ctx.Done():
			assert.Fail(t, "timeout")
		}
	})

	t.Run(LangJSONata, func(t *testing.T) {
		n, _ := NewSnippetNode(SnippetNodeConfig{
			Lang: LangJSONata,
			Code: "$",
		})
		defer func() { _ = n.Close() }()

		io := port.New()
		ioPort, _ := n.Port(node.PortIO)
		ioPort.Link(io)

		proc := process.New()
		defer proc.Exit(nil)

		ioStream := io.Open(proc)

		inPayload := primitive.NewString(faker.UUIDHyphenated())
		inPck := packet.New(inPayload)

		ioStream.Send(inPck)

		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()

		select {
		case outPck := <-ioStream.Receive():
			assert.Equal(t, inPayload, outPck.Payload())
		case <-ctx.Done():
			assert.Fail(t, "timeout")
		}
	})
}

func BenchmarkSnippetNode_Send(b *testing.B) {
	b.Run(LangTypescript, func(b *testing.B) {
		n, _ := NewSnippetNode(SnippetNodeConfig{
			Lang: LangTypescript,
			Code: `
function main(inPayload: any): any {
	return inPayload;
}
				`,
		})
		defer func() { _ = n.Close() }()

		io := port.New()
		ioPort, _ := n.Port(node.PortIO)
		ioPort.Link(io)

		for i := 0; i < b.N; i++ {
			proc := process.New()
			defer proc.Exit(nil)

			ioStream := io.Open(proc)

			inPayload := primitive.NewString(faker.UUIDHyphenated())
			inPck := packet.New(inPayload)

			ioStream.Send(inPck)

			ctx, cancel := context.WithTimeout(context.Background(), time.Second)
			defer cancel()

			select {
			case <-ioStream.Receive():
			case <-ctx.Done():
				assert.Fail(b, "timeout")
			}
		}
	})

	b.Run(LangJavascript, func(b *testing.B) {
		n, _ := NewSnippetNode(SnippetNodeConfig{
			Lang: LangJavascript,
			Code: `
function main(inPayload) {
	return inPayload;
}
				`,
		})
		defer func() { _ = n.Close() }()

		io := port.New()
		ioPort, _ := n.Port(node.PortIO)
		ioPort.Link(io)

		for i := 0; i < b.N; i++ {
			proc := process.New()
			defer proc.Exit(nil)

			ioStream := io.Open(proc)

			inPayload := primitive.NewString(faker.UUIDHyphenated())
			inPck := packet.New(inPayload)

			ioStream.Send(inPck)

			ctx, cancel := context.WithTimeout(context.Background(), time.Second)
			defer cancel()

			select {
			case <-ioStream.Receive():
			case <-ctx.Done():
				assert.Fail(b, "timeout")
			}
		}
	})

	b.Run(LangJSON, func(b *testing.B) {
		n, _ := NewSnippetNode(SnippetNodeConfig{
			Lang: LangJSON,
			Code: fmt.Sprintf("\"%s\"", faker.UUIDHyphenated()),
		})
		defer func() { _ = n.Close() }()

		io := port.New()
		ioPort, _ := n.Port(node.PortIO)
		ioPort.Link(io)

		for i := 0; i < b.N; i++ {
			proc := process.New()
			defer proc.Exit(nil)

			ioStream := io.Open(proc)

			inPayload := primitive.NewString(faker.UUIDHyphenated())
			inPck := packet.New(inPayload)

			ioStream.Send(inPck)

			ctx, cancel := context.WithTimeout(context.Background(), time.Second)
			defer cancel()

			select {
			case <-ioStream.Receive():
			case <-ctx.Done():
				assert.Fail(b, "timeout")
			}
		}
	})

	b.Run(LangJSONata, func(b *testing.B) {
		n, _ := NewSnippetNode(SnippetNodeConfig{
			Lang: LangJSONata,
			Code: "$",
		})
		defer func() { _ = n.Close() }()

		io := port.New()
		ioPort, _ := n.Port(node.PortIO)
		ioPort.Link(io)

		for i := 0; i < b.N; i++ {
			proc := process.New()
			defer proc.Exit(nil)

			ioStream := io.Open(proc)

			inPayload := primitive.NewString(faker.UUIDHyphenated())
			inPck := packet.New(inPayload)

			ioStream.Send(inPck)

			ctx, cancel := context.WithTimeout(context.Background(), time.Second)
			defer cancel()

			select {
			case <-ioStream.Receive():
			case <-ctx.Done():
				assert.Fail(b, "timeout")
			}
		}
	})
}
