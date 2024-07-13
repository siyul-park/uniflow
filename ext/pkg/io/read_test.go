package io

import (
	"bytes"
	"context"
	"io"
	"os"
	"testing"
	"time"

	"github.com/go-faker/faker/v4"
	"github.com/siyul-park/uniflow/pkg/node"
	"github.com/siyul-park/uniflow/pkg/packet"
	"github.com/siyul-park/uniflow/pkg/port"
	"github.com/siyul-park/uniflow/pkg/process"
	"github.com/siyul-park/uniflow/pkg/types"
	"github.com/stretchr/testify/assert"
)

func TestNewReadNode(t *testing.T) {
	n := NewReadNode(NewOsFs())
	assert.NotNil(t, n)
	assert.NoError(t, n.Close())
}

func TestReadNode_SendAndReceive(t *testing.T) {
	t.Run("Static", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.TODO(), time.Second)
		defer cancel()

		data := []byte(faker.Sentence())

		buf := bytes.NewBuffer(data)
		fs := OpenFileFunc(func(name string, flag int, perm os.FileMode) (io.ReadWriteCloser, error) {
			return &nopReadWriteCloser{buf}, nil
		})

		n := NewReadNode(fs)
		defer n.Close()

		err := n.Open("")
		assert.NoError(t, err)

		in := port.NewOut()
		in.Link(n.In(node.PortIn))

		proc := process.New()
		defer proc.Exit(nil)

		inWriter := in.Open(proc)

		inPayload := types.NewInt(len(data))
		inPck := packet.New(inPayload)

		inWriter.Write(inPck)

		select {
		case outPck := <-inWriter.Receive():
			assert.Equal(t, types.NewString(string(data)), outPck.Payload())
		case <-ctx.Done():
			assert.Fail(t, ctx.Err().Error())
		}
	})

	t.Run("Dynamic", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.TODO(), time.Second)
		defer cancel()

		data := []byte(faker.Sentence())

		buf := bytes.NewBuffer(data)
		fs := OpenFileFunc(func(name string, flag int, perm os.FileMode) (io.ReadWriteCloser, error) {
			return &nopReadWriteCloser{buf}, nil
		})

		n := NewReadNode(fs)
		defer n.Close()

		in := port.NewOut()
		in.Link(n.In(node.PortIn))

		proc := process.New()
		defer proc.Exit(nil)

		inWriter := in.Open(proc)

		inPayload := types.NewSlice(
			types.NewString(""),
			types.NewInt(len(data)),
		)
		inPck := packet.New(inPayload)

		inWriter.Write(inPck)

		select {
		case outPck := <-inWriter.Receive():
			assert.Equal(t, types.NewString(string(data)), outPck.Payload())
		case <-ctx.Done():
			assert.Fail(t, ctx.Err().Error())
		}
	})
}

func TestReadNodeCodec_Decode(t *testing.T) {
	codec := NewReadNodeCodec()

	spec := &ReadNodeSpec{
		Filename: "stdin",
	}

	n, err := codec.Compile(spec)
	assert.NoError(t, err)
	assert.NotNil(t, n)
	assert.NoError(t, n.Close())
}
