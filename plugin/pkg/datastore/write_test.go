package datastore

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/go-faker/faker/v4"
	"github.com/siyul-park/uniflow/pkg/node"
	"github.com/siyul-park/uniflow/pkg/object"
	"github.com/siyul-park/uniflow/pkg/packet"
	"github.com/siyul-park/uniflow/pkg/port"
	"github.com/siyul-park/uniflow/pkg/process"
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

	f, _ := os.CreateTemp("", "*")
	defer f.Close()

	n := NewWriteNode(f)
	defer n.Close()

	in := port.NewOut()
	in.Link(n.In(node.PortIn))

	proc := process.New()
	defer proc.Exit(nil)

	inWriter := in.Open(proc)

	inPayload := object.NewString(faker.UUIDHyphenated())
	inPck := packet.New(inPayload)

	inWriter.Write(inPck)

	select {
	case outPck := <-inWriter.Receive():
		assert.Equal(t, int64(inPayload.Len()), outPck.Payload().Interface())
	case <-ctx.Done():
		assert.Fail(t, ctx.Err().Error())
	}
}

func TestWriteNodeCodec_Decode(t *testing.T) {
	codec := NewWriteNodeCodec()

	spec := &WriteNodeSpec{
		File: "stdout",
	}

	n, err := codec.Decode(spec)
	assert.NoError(t, err)
	assert.NotNil(t, n)
	assert.NoError(t, n.Close())
}
