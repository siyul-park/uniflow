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

func TestNewWriteNode(t *testing.T) {
	n := NewWriteNode(NewOsFs())
	assert.NotNil(t, n)
	assert.NoError(t, n.Close())
}

func TestWriteNode_SendAndReceive(t *testing.T) {
	t.Run("Static", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.TODO(), time.Second)
		defer cancel()

		buf := bytes.NewBuffer(nil)
		fs := OpenFileFunc(func(name string, flag int, perm os.FileMode) (io.ReadWriteCloser, error) {
			return &nopReadWriteCloser{buf}, nil
		})

		n := NewWriteNode(fs)
		defer n.Close()

		err := n.Open("")
		assert.NoError(t, err)

		in := port.NewOut()
		in.Link(n.In(node.PortIn))

		proc := process.New()
		defer proc.Exit(nil)

		inWriter := in.Open(proc)

		data := faker.UUIDHyphenated()

		inPayload := types.NewString(data)
		inPck := packet.New(inPayload)

		inWriter.Write(inPck)

		select {
		case outPck := <-inWriter.Receive():
			assert.Equal(t, types.NewInt64(int64(len(data))), outPck.Payload())
			assert.Equal(t, buf.String(), data)
		case <-ctx.Done():
			assert.Fail(t, ctx.Err().Error())
		}
	})

	t.Run("Dynamic", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.TODO(), time.Second)
		defer cancel()

		buf := bytes.NewBuffer(nil)
		fs := OpenFileFunc(func(name string, flag int, perm os.FileMode) (io.ReadWriteCloser, error) {
			return &nopReadWriteCloser{buf}, nil
		})

		n := NewWriteNode(fs)
		defer n.Close()

		in := port.NewOut()
		in.Link(n.In(node.PortIn))

		proc := process.New()
		defer proc.Exit(nil)

		inWriter := in.Open(proc)

		data := faker.UUIDHyphenated()

		inPayload := types.NewSlice(
			types.NewString(""),
			types.NewString(data),
		)
		inPck := packet.New(inPayload)

		inWriter.Write(inPck)

		select {
		case outPck := <-inWriter.Receive():
			assert.Equal(t, types.NewInt64(int64(len(data))), outPck.Payload())
			assert.Equal(t, buf.String(), data)
		case <-ctx.Done():
			assert.Fail(t, ctx.Err().Error())
		}
	})
}

func TestWriteNodeCodec_Decode(t *testing.T) {
	codec := NewWriteNodeCodec()

	spec := &WriteNodeSpec{
		Filename: "stdout",
	}

	n, err := codec.Compile(spec)
	assert.NoError(t, err)
	assert.NotNil(t, n)
	assert.NoError(t, n.Close())
}
