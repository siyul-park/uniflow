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

func TestNewWriteNode(t *testing.T) {
	n := NewWriteNode(NewOSFileSystem())
	assert.NotNil(t, n)
	assert.NoError(t, n.Close())
}

func TestWriteNode_SendAndReceive(t *testing.T) {
	t.Run("Static", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.TODO(), time.Second)
		defer cancel()

		buf := bytes.NewBuffer(nil)
		fs := FileOpenFunc(func(name string, flag int, perm os.FileMode) (io.ReadWriteCloser, error) {
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
		fs := FileOpenFunc(func(name string, flag int, perm os.FileMode) (io.ReadWriteCloser, error) {
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

func BenchmarkWriteNode_SendAndReceive(b *testing.B) {
	buf := bytes.NewBuffer(nil)
	fs := FileOpenFunc(func(name string, flag int, perm os.FileMode) (io.ReadWriteCloser, error) {
		return &nopReadWriteCloser{buf}, nil
	})

	n := NewWriteNode(fs)
	defer n.Close()

	n.Open("")

	in := port.NewOut()
	in.Link(n.In(node.PortIn))

	proc := process.New()
	defer proc.Exit(nil)

	inWriter := in.Open(proc)

	data := faker.UUIDHyphenated()

	inPayload := types.NewString(data)
	inPck := packet.New(inPayload)

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		inWriter.Write(inPck)
		<-inWriter.Receive()
	}
}
