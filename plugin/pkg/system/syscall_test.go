package system

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

func TestNewSyscallNode(t *testing.T) {
	n, err := NewSyscallNode(func() {})

	assert.NoError(t, err)
	assert.NotNil(t, n)

	assert.NoError(t, n.Close())
}

func TestSyscallNode_SetArguments(t *testing.T) {
	t.Run(language.Text, func(t *testing.T) {
		n, _ := NewSyscallNode(func(arg any) any { return arg })
		defer n.Close()

		err := n.SetArguments(language.Text, "foo")
		assert.NoError(t, err)
	})

	t.Run(language.Typescript, func(t *testing.T) {
		n, _ := NewSyscallNode(func(arg any) any { return arg })
		defer n.Close()

		err := n.SetArguments(language.Typescript, "$")
		assert.NoError(t, err)
	})

	t.Run(language.Javascript, func(t *testing.T) {
		n, _ := NewSyscallNode(func(arg any) any { return arg })
		defer n.Close()

		err := n.SetArguments(language.Javascript, "$")
		assert.NoError(t, err)
	})

	t.Run(language.JSON, func(t *testing.T) {
		n, _ := NewSyscallNode(func(arg any) any { return arg })
		defer n.Close()

		err := n.SetArguments(language.JSON, "\"foo\"")
		assert.NoError(t, err)
	})

	t.Run(language.JSONata, func(t *testing.T) {
		n, _ := NewSyscallNode(func(arg any) any { return arg })
		defer n.Close()

		err := n.SetArguments(language.JSONata, "$")
		assert.NoError(t, err)
	})

	t.Run(language.YAML, func(t *testing.T) {
		n, _ := NewSyscallNode(func(arg any) any { return arg })
		defer n.Close()

		err := n.SetArguments(language.YAML, "\"foo\"")
		assert.NoError(t, err)
	})
}

func TestSyscallNode_SendAndReceive(t *testing.T) {
	t.Run(language.Text, func(t *testing.T) {
		n, _ := NewSyscallNode(func(arg any) any { return arg })
		defer n.Close()

		_ = n.SetArguments(language.Text, "foo")

		io := port.New()
		ioPort := n.Port(node.PortIO)
		ioPort.Link(io)

		proc := process.New()
		defer proc.Exit(nil)

		ioStream := io.Open(proc)

		inPayload := primitive.NewString(faker.UUIDHyphenated())
		inPck := packet.New(inPayload)

		ioStream.Send(inPck)

		ctx, cancel := context.WithTimeout(context.TODO(), time.Second)
		defer cancel()

		select {
		case outPck := <-ioStream.Receive():
			assert.Equal(t, primitive.NewString("foo"), outPck.Payload())
		case <-ctx.Done():
			assert.Fail(t, ctx.Err().Error())
		}
	})

	t.Run(language.Typescript, func(t *testing.T) {
		n, _ := NewSyscallNode(func(arg any) any { return arg })
		defer n.Close()

		_ = n.SetArguments(language.Typescript, "$")

		io := port.New()
		ioPort := n.Port(node.PortIO)
		ioPort.Link(io)

		proc := process.New()
		defer proc.Exit(nil)

		ioStream := io.Open(proc)

		inPayload := primitive.NewString(faker.UUIDHyphenated())
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

	t.Run(language.Javascript, func(t *testing.T) {
		n, _ := NewSyscallNode(func(arg any) any { return arg })
		defer n.Close()

		_ = n.SetArguments(language.Javascript, "$")

		io := port.New()
		ioPort := n.Port(node.PortIO)
		ioPort.Link(io)

		proc := process.New()
		defer proc.Exit(nil)

		ioStream := io.Open(proc)

		inPayload := primitive.NewString(faker.UUIDHyphenated())
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
		n, _ := NewSyscallNode(func(arg any) any { return arg })
		defer n.Close()

		_ = n.SetArguments(language.JSON, "\"foo\"")

		io := port.New()
		ioPort := n.Port(node.PortIO)
		ioPort.Link(io)

		proc := process.New()
		defer proc.Exit(nil)

		ioStream := io.Open(proc)

		inPayload := primitive.NewString(faker.UUIDHyphenated())
		inPck := packet.New(inPayload)

		ioStream.Send(inPck)

		ctx, cancel := context.WithTimeout(context.TODO(), time.Second)
		defer cancel()

		select {
		case outPck := <-ioStream.Receive():
			assert.Equal(t, primitive.NewString("foo"), outPck.Payload())
		case <-ctx.Done():
			assert.Fail(t, ctx.Err().Error())
		}
	})

	t.Run(language.JSONata, func(t *testing.T) {
		n, _ := NewSyscallNode(func(arg any) any { return arg })
		defer n.Close()

		_ = n.SetArguments(language.JSONata, "$")

		io := port.New()
		ioPort := n.Port(node.PortIO)
		ioPort.Link(io)

		proc := process.New()
		defer proc.Exit(nil)

		ioStream := io.Open(proc)

		inPayload := primitive.NewString(faker.UUIDHyphenated())
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
