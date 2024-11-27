package io

import (
	"bytes"
	"context"
	"io"
	"testing"
	"time"

	"github.com/siyul-park/uniflow/pkg/node"
	"github.com/siyul-park/uniflow/pkg/packet"
	"github.com/siyul-park/uniflow/pkg/port"
	"github.com/siyul-park/uniflow/pkg/process"
	"github.com/siyul-park/uniflow/pkg/types"
	"github.com/stretchr/testify/assert"
)

func TestPrintNodeCodec_Compile(t *testing.T) {
	t.Run("static", func(t *testing.T) {
		codec := NewPrintNodeCodec(FileOpenFunc(func(name string, flag int) (io.ReadWriteCloser, error) {
			return &nopReadWriteCloser{bytes.NewBuffer(nil)}, nil
		}))

		spec := &PrintNodeSpec{
			Filename: "output.txt",
		}

		n, err := codec.Compile(spec)
		assert.NoError(t, err)
		assert.NotNil(t, n)
		assert.NoError(t, n.Close())
	})

	t.Run("dynamic", func(t *testing.T) {
		codec := NewPrintNodeCodec(FileOpenFunc(func(name string, flag int) (io.ReadWriteCloser, error) {
			return &nopReadWriteCloser{bytes.NewBuffer(nil)}, nil
		}))

		spec := &PrintNodeSpec{}

		n, err := codec.Compile(spec)
		assert.NoError(t, err)
		assert.NotNil(t, n)
		assert.NoError(t, n.Close())
	})
}

func TestNewPrintNode(t *testing.T) {
	buf := &bytes.Buffer{}
	n := NewPrintNode(&nopReadWriteCloser{ReadWriter: buf})
	assert.NotNil(t, n)
	assert.NoError(t, n.Close())
}

func TestPrintNode_SendAndReceive(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.TODO(), time.Second)
	defer cancel()

	buf := &bytes.Buffer{}

	n := NewPrintNode(&nopReadWriteCloser{ReadWriter: buf})
	defer n.Close()

	in := port.NewOut()
	in.Link(n.In(node.PortIn))

	proc := process.New()
	defer proc.Exit(nil)

	inWriter := in.Open(proc)

	inPayload := types.NewSlice(
		types.NewString("%s %d"),
		types.NewString("hello"),
		types.NewInt(123),
	)
	inPck := packet.New(inPayload)
	inWriter.Write(inPck)

	select {
	case <-inWriter.Receive():
		assert.Equal(t, "hello 123", buf.String())
	case <-ctx.Done():
		assert.Fail(t, ctx.Err().Error())
	}
}

func TestNewDynPrintNode(t *testing.T) {
	fs := FileOpenFunc(func(name string, flag int) (io.ReadWriteCloser, error) {
		return &nopReadWriteCloser{bytes.NewBuffer(nil)}, nil
	})
	n := NewDynPrintNode(fs)
	assert.NotNil(t, n)
	assert.NoError(t, n.Close())
}

func TestDynPrintNode_SendAndReceive(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.TODO(), time.Second)
	defer cancel()

	buf := bytes.NewBuffer(nil)
	fs := FileOpenFunc(func(name string, flag int) (io.ReadWriteCloser, error) {
		return &nopReadWriteCloser{buf}, nil
	})

	n := NewDynPrintNode(fs)
	defer n.Close()

	in := port.NewOut()
	in.Link(n.In(node.PortIn))

	proc := process.New()
	defer proc.Exit(nil)

	inWriter := in.Open(proc)

	inPayload := types.NewSlice(
		types.NewString(""),
		types.NewString("%s %d"),
		types.NewString("hello"),
		types.NewInt(123),
	)
	inPck := packet.New(inPayload)
	inWriter.Write(inPck)

	select {
	case <-inWriter.Receive():
		assert.Equal(t, "hello 123", buf.String())
	case <-ctx.Done():
		assert.Fail(t, ctx.Err().Error())
	}
}
