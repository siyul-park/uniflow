package system

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
	"github.com/siyul-park/uniflow/plugin/internal/language"
	"github.com/stretchr/testify/assert"
)

func TestNewBridgeNode(t *testing.T) {
	n, err := NewBridgeNode(func() {})

	assert.NoError(t, err)
	assert.NotNil(t, n)

	assert.NoError(t, n.Close())
}

func TestBridgeNode_SendAndReceive(t *testing.T) {
	t.Run("Arguments, Returns = 0", func(t *testing.T) {
		n, _ := NewBridgeNode(func() {})
		defer n.Close()

		io := port.NewOut()
		io.Link(n.In(node.PortIO))

		proc := process.New()
		defer proc.Close()

		ioWriter := io.Open(proc)

		inPayload := primitive.NewString(faker.UUIDHyphenated())
		inPck := packet.New(inPayload)

		ioWriter.Write(inPck)

		ctx, cancel := context.WithTimeout(context.TODO(), time.Second)
		defer cancel()

		select {
		case outPck := <-ioWriter.Receive():
			assert.Nil(t, outPck.Payload())
		case <-ctx.Done():
			assert.Fail(t, ctx.Err().Error())
		}
	})

	t.Run("Arguments = 1, Returns == 1", func(t *testing.T) {
		n, _ := NewBridgeNode(func(arg any) any {
			return arg
		})
		defer n.Close()

		_ = n.SetArguments("$")

		io := port.NewOut()
		io.Link(n.In(node.PortIO))

		proc := process.New()
		defer proc.Close()

		ioWriter := io.Open(proc)

		inPayload := primitive.NewString(faker.UUIDHyphenated())
		inPck := packet.New(inPayload)

		ioWriter.Write(inPck)

		ctx, cancel := context.WithTimeout(context.TODO(), time.Second)
		defer cancel()

		select {
		case outPck := <-ioWriter.Receive():
			assert.Equal(t, inPayload, outPck.Payload())
		case <-ctx.Done():
			assert.Fail(t, ctx.Err().Error())
		}
	})

	t.Run("Arguments > 1, Returns == 1", func(t *testing.T) {
		n, _ := NewBridgeNode(func(arg1, arg2 any) any {
			return arg2
		})
		defer n.Close()

		_ = n.SetArguments("$", "$")

		io := port.NewOut()
		io.Link(n.In(node.PortIO))

		proc := process.New()
		defer proc.Close()

		ioWriter := io.Open(proc)

		inPayload := primitive.NewString(faker.UUIDHyphenated())
		inPck := packet.New(inPayload)

		ioWriter.Write(inPck)

		ctx, cancel := context.WithTimeout(context.TODO(), time.Second)
		defer cancel()

		select {
		case outPck := <-ioWriter.Receive():
			assert.Equal(t, inPayload, outPck.Payload())
		case <-ctx.Done():
			assert.Fail(t, ctx.Err().Error())
		}
	})

	t.Run("Arguments == Context, Returns == 1", func(t *testing.T) {
		n, _ := NewBridgeNode(func(ctx context.Context, arg any) any {
			return arg
		})
		defer n.Close()

		io := port.NewOut()
		io.Link(n.In(node.PortIO))

		proc := process.New()
		defer proc.Close()

		ioWriter := io.Open(proc)

		inPayload := primitive.NewString(faker.UUIDHyphenated())
		inPck := packet.New(inPayload)

		ioWriter.Write(inPck)

		ctx, cancel := context.WithTimeout(context.TODO(), time.Second)
		defer cancel()

		select {
		case outPck := <-ioWriter.Receive():
			assert.Equal(t, inPayload, outPck.Payload())
		case <-ctx.Done():
			assert.Fail(t, ctx.Err().Error())
		}
	})

	t.Run("Arguments == 1, Returns > 2", func(t *testing.T) {
		n, _ := NewBridgeNode(func(arg any) (any, any) {
			return arg, arg
		})
		defer n.Close()

		_ = n.SetArguments("$")

		io := port.NewOut()
		io.Link(n.In(node.PortIO))

		proc := process.New()
		defer proc.Close()

		ioWriter := io.Open(proc)

		inPayload := primitive.NewString(faker.UUIDHyphenated())
		inPck := packet.New(inPayload)

		ioWriter.Write(inPck)

		ctx, cancel := context.WithTimeout(context.TODO(), time.Second)
		defer cancel()

		select {
		case outPck := <-ioWriter.Receive():
			assert.Equal(t, primitive.NewSlice(inPayload, inPayload), outPck.Payload())
		case <-ctx.Done():
			assert.Fail(t, ctx.Err().Error())
		}
	})

	t.Run("Arguments == 1, Returns == error", func(t *testing.T) {
		n, _ := NewBridgeNode(func(arg any) error {
			return fmt.Errorf("%v", arg)
		})
		defer n.Close()

		_ = n.SetArguments("$")

		io := port.NewOut()
		io.Link(n.In(node.PortIO))

		err := port.NewIn()
		n.Out(node.PortErr).Link(err)

		proc := process.New()
		defer proc.Close()

		ioWriter := io.Open(proc)
		errReader := err.Open(proc)

		inPayload := primitive.NewString(faker.UUIDHyphenated())
		inPck := packet.New(inPayload)

		ioWriter.Write(inPck)

		ctx, cancel := context.WithTimeout(context.TODO(), time.Second)
		defer cancel()

		select {
		case outPck := <-errReader.Read():
			assert.NotNil(t, outPck)
			errReader.Receive(outPck)
		case <-ctx.Done():
			assert.Fail(t, ctx.Err().Error())
		}

		select {
		case backPck := <-ioWriter.Receive():
			assert.NotNil(t, backPck)
		case <-ctx.Done():
			assert.Fail(t, "timeout")
		}
	})
}

func TestBridgeNodeCodec_Decode(t *testing.T) {
	table := NewBridgeTable()

	operation := faker.UUIDHyphenated()

	table.Store(operation, func(arg any) any {
		return arg
	})

	codec := NewBridgeNodeCodec(table)

	spec := &BridgeNodeSpec{
		Opcode:    operation,
		Lang:      language.Text,
		Arguments: []string{"foo"},
	}
	n, err := codec.Decode(spec)
	assert.NoError(t, err)
	assert.NotNil(t, n)

	assert.NoError(t, n.Close())
}
