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
	"github.com/stretchr/testify/require"
)

func TestScanNodeCodec_Compile(t *testing.T) {
	t.Run("static", func(t *testing.T) {
		codec := NewScanNodeCodec(FileOpenFunc(func(name string, flag int) (io.ReadWriteCloser, error) {
			return &nopReadWriteCloser{bytes.NewBuffer(nil)}, nil
		}))

		spec := &ScanNodeSpec{
			Filename: "stdin",
		}

		n, err := codec.Compile(spec)
		require.NoError(t, err)
		require.NotNil(t, n)
		require.NoError(t, n.Close())
	})

	t.Run("dynamic", func(t *testing.T) {
		codec := NewScanNodeCodec(FileOpenFunc(func(name string, flag int) (io.ReadWriteCloser, error) {
			return &nopReadWriteCloser{bytes.NewBuffer(nil)}, nil
		}))

		spec := &ScanNodeSpec{}

		n, err := codec.Compile(spec)
		require.NoError(t, err)
		require.NotNil(t, n)
		require.NoError(t, n.Close())
	})
}

func TestNewScanNode(t *testing.T) {
	n := NewScanNode(&nopReadWriteCloser{bytes.NewBuffer(nil)})
	require.NotNil(t, n)
	require.NoError(t, n.Close())
}

func TestScanNode_SendAndReceive(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.TODO(), time.Second)
	defer cancel()

	buf := bytes.NewBuffer([]byte("true 3.14 42 HelloWorld 123"))

	n := NewScanNode(&nopReadWriteCloser{buf})
	defer n.Close()

	in := port.NewOut()
	in.Link(n.In(node.PortIn))

	proc := process.New()
	defer proc.Exit(nil)

	inWriter := in.Open(proc)

	inPayload := types.NewString("%t %f %d %s %c")
	inPck := packet.New(inPayload)

	inWriter.Write(inPck)

	select {
	case outPck := <-inWriter.Receive():
		require.Equal(t, types.KindSlice, outPck.Payload().Kind())
	case <-ctx.Done():
		require.Fail(t, ctx.Err().Error())
	}
}

func TestNewDynScanNode(t *testing.T) {
	n := NewDynScanNode(NewOSFileSystem())
	require.NotNil(t, n)
	require.NoError(t, n.Close())
}

func TestDynScanNode_SendAndReceive(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.TODO(), time.Second)
	defer cancel()

	buf := bytes.NewBuffer([]byte("true 3.14 42 HelloWorld 123"))
	fs := FileOpenFunc(func(name string, flag int) (io.ReadWriteCloser, error) {
		return &nopReadWriteCloser{buf}, nil
	})

	n := NewDynScanNode(fs)
	defer n.Close()

	in := port.NewOut()
	in.Link(n.In(node.PortIn))

	proc := process.New()
	defer proc.Exit(nil)

	inWriter := in.Open(proc)

	inPayload := types.NewSlice(
		types.NewString(""),
		types.NewString("%t %f %d %s %c"),
	)
	inPck := packet.New(inPayload)

	inWriter.Write(inPck)

	select {
	case outPck := <-inWriter.Receive():
		require.Equal(t, types.KindSlice, outPck.Payload().Kind())
	case <-ctx.Done():
		require.Fail(t, ctx.Err().Error())
	}
}
