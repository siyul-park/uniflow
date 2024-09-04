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

func TestReadNodeCodec_Decode(t *testing.T) {
	codec := NewReadNodeCodec(NewOSFileSystem())

	spec := &ReadNodeSpec{
		Filename: "stdin",
	}

	n, err := codec.Compile(spec)
	assert.NoError(t, err)
	assert.NotNil(t, n)
	assert.NoError(t, n.Close())
}

func TestNewReadNode(t *testing.T) {
	n := NewReadNode(NewOSFileSystem())
	assert.NotNil(t, n)
	assert.NoError(t, n.Close())
}

func TestReadNode_SendAndReceive(t *testing.T) {
	t.Run("Static", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.TODO(), time.Second)
		defer cancel()

		data := []byte(faker.Sentence())

		buf := bytes.NewBuffer(data)
		fs := FileOpenFunc(func(name string, flag int, perm os.FileMode) (io.ReadWriteCloser, error) {
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
		fs := FileOpenFunc(func(name string, flag int, perm os.FileMode) (io.ReadWriteCloser, error) {
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

func BenchmarkReadNode_SendAndReceive(b *testing.B) {
	data := []byte(faker.Sentence())

	buf := bytes.NewBuffer(data)
	fs := FileOpenFunc(func(name string, flag int, perm os.FileMode) (io.ReadWriteCloser, error) {
		return &nopReadWriteCloser{buf}, nil
	})

	n := NewReadNode(fs)
	defer n.Close()

	n.Open("")

	in := port.NewOut()
	in.Link(n.In(node.PortIn))

	proc := process.New()
	defer proc.Exit(nil)

	inWriter := in.Open(proc)

	inPayload := types.NewInt(len(data))
	inPck := packet.New(inPayload)

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		inWriter.Write(inPck)
		<-inWriter.Receive()
	}
}
