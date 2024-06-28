package control

import (
	"context"
	"testing"
	"time"

	"github.com/go-faker/faker/v4"
	"github.com/siyul-park/uniflow/node"
	"github.com/siyul-park/uniflow/object"
	"github.com/siyul-park/uniflow/packet"
	"github.com/siyul-park/uniflow/port"
	"github.com/siyul-park/uniflow/process"
	"github.com/stretchr/testify/assert"
)

func TestNewNOPNode(t *testing.T) {
	n := NewNOPNode()
	assert.NotNil(t, n)
	assert.NoError(t, n.Close())
}

func TestNOPNode_Port(t *testing.T) {
	n := NewNOPNode()
	defer n.Close()

	assert.NotNil(t, n.In(node.PortIn))
}

func TestNOPNode_SendAndReceive(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.TODO(), time.Second)
	defer cancel()

	n := NewNOPNode()
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
	case backPck := <-inWriter.Receive():
		assert.Equal(t, packet.None, backPck)
	case <-ctx.Done():
		assert.Fail(t, "timeout")
	}
}

func TestNOPNodeCodec_Decode(t *testing.T) {
	codec := NewNOPNodeCodec()

	spec := &NOPNodeSpec{}

	n, err := codec.Decode(spec)
	assert.NoError(t, err)
	assert.NotNil(t, n)
	assert.NoError(t, n.Close())
}
