package io

import (
	"bytes"
	"context"
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
	f, _ := os.CreateTemp("", "*")
	defer f.Close()

	n := NewWriteNode(f)
	assert.NotNil(t, n)
	assert.NoError(t, n.Close())
}

func TestWriteNode_SendAndReceive(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.TODO(), time.Second)
	defer cancel()

	buf := bytes.NewBuffer(nil)

	n := NewWriteNode(&nopReadWriteCloser{buf})
	defer n.Close()

	in := port.NewOut()
	in.Link(n.In(node.PortIn))

	proc := process.New()
	defer proc.Exit(nil)

	inWriter := in.Open(proc)

	inPayload := types.NewString(faker.UUIDHyphenated())
	inPck := packet.New(inPayload)

	inWriter.Write(inPck)

	select {
	case outPck := <-inWriter.Receive():
		assert.Equal(t, types.NewInt64(int64(inPayload.Len())), outPck.Payload())
		assert.Equal(t, types.NewString(buf.String()), inPayload)
	case <-ctx.Done():
		assert.Fail(t, ctx.Err().Error())
	}
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
